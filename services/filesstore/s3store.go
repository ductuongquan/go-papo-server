// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package filesstore

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"

	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
)

type S3FileBackend struct {
	endpoint   string
	accessKey  string
	secretKey  string
	secure     bool
	signV2     bool
	region     string
	bucket     string
	pathPrefix string
	encrypt    bool
	trace      bool
}

const (
	// This is not exported by minio. See: https://github.com/minio/minio-go/issues/1339
	bucketNotFound = "NoSuchBucket"
)

// Similar to s3.New() but allows initialization of signature v2 or signature v4 client.
// If signV2 input is false, function always returns signature v4.
//
// Additionally this function also takes a user defined region, if set
// disables automatic region lookup.
func (b *S3FileBackend) s3New() (*s3.Client, error) {
	var creds *credentials.Credentials

	if b.accessKey == "" && b.secretKey == "" {
		creds = credentials.NewIAM("")
	} else if b.signV2 {
		creds = credentials.NewStatic(b.accessKey, b.secretKey, "", credentials.SignatureV2)
	} else {
		creds = credentials.NewStatic(b.accessKey, b.secretKey, "", credentials.SignatureV4)
	}

	opts := s3.Options{
		Creds:  creds,
		Secure: b.secure,
		Region: b.region,
	}
	s3Clnt, err := s3.New(b.endpoint, &opts)
	if err != nil {
		return nil, err
	}

	if b.trace {
		s3Clnt.TraceOn(os.Stdout)
	}

	return s3Clnt, nil
}

func (b *S3FileBackend) TestConnection() *model.AppError {
	s3Clnt, err := b.s3New()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	exists := true
	// If a path prefix is present, we attempt to test the bucket by listing objects under the path
	// and just checking the first response. This is because the BucketExists call is only at a bucket level
	// and sometimes the user might only be allowed access to the specified path prefix.
	if b.pathPrefix != "" {
		obj := <-s3Clnt.ListObjects(context.Background(), b.bucket, s3.ListObjectsOptions{Prefix: b.pathPrefix})
		if obj.Err != nil {
			typedErr := s3.ToErrorResponse(obj.Err)
			if typedErr.Code != bucketNotFound {
				return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.list_objects.app_error", nil, obj.Err.Error(), http.StatusInternalServerError)
			}
			exists = false
		}
	} else {
		exists, err = s3Clnt.BucketExists(context.Background(), b.bucket)
		if err != nil {
			return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.bucket_exists.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if exists {
		mlog.Debug("Connection to S3 or minio is good. Bucket exists.")
	} else {
		mlog.Warn("Bucket specified does not exist. Attempting to create...")
		err := s3Clnt.MakeBucket(context.Background(), b.bucket, s3.MakeBucketOptions{Region: b.region})
		if err != nil {
			return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.bucket_create.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return nil
}

// Caller must close the first return value
func (b *S3FileBackend) Reader(path string) (ReadCloseSeeker, *model.AppError) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, model.NewAppError("Reader", "api.file.reader.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	minioObject, err := s3Clnt.GetObject(context.Background(), b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, model.NewAppError("Reader", "api.file.reader.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return minioObject, nil
}

func (b *S3FileBackend) ReadFile(path string) ([]byte, *model.AppError) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	minioObject, err := s3Clnt.GetObject(context.Background(), b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	defer minioObject.Close()
	if f, err := ioutil.ReadAll(minioObject); err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return f, nil
	}
}

func (b *S3FileBackend) FileExists(path string) (bool, *model.AppError) {
	s3Clnt, err := b.s3New()

	if err != nil {
		return false, model.NewAppError("FileExists", "api.file.file_exists.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	_, err = s3Clnt.StatObject(context.Background(), b.bucket, path, s3.StatObjectOptions{})

	if err == nil {
		return true, nil
	}

	var s3Err s3.ErrorResponse
	if errors.As(err, &s3Err); s3Err.Code == "NoSuchKey" {
		return false, nil
	}

	return false, model.NewAppError("FileExists", "api.file.file_exists.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
}

func (b *S3FileBackend) CopyFile(oldPath, newPath string) *model.AppError {
	s3Clnt, err := b.s3New()
	if err != nil {
		return model.NewAppError("copyFile", "api.file.write_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	oldPath = filepath.Join(b.pathPrefix, oldPath)
	newPath = filepath.Join(b.pathPrefix, newPath)
	srcOpts := s3.CopySrcOptions{
		Bucket:     b.bucket,
		Object:     oldPath,
		Encryption: encrypt.NewSSE(),
	}
	dstOpts := s3.CopyDestOptions{
		Bucket:     b.bucket,
		Object:     newPath,
		Encryption: encrypt.NewSSE(),
	}
	if _, err = s3Clnt.CopyObject(context.Background(), dstOpts, srcOpts); err != nil {
		return model.NewAppError("copyFile", "api.file.move_file.copy_within_s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (b *S3FileBackend) MoveFile(oldPath, newPath string) *model.AppError {
	s3Clnt, err := b.s3New()
	if err != nil {
		return model.NewAppError("moveFile", "api.file.write_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	oldPath = filepath.Join(b.pathPrefix, oldPath)
	newPath = filepath.Join(b.pathPrefix, newPath)
	srcOpts := s3.CopySrcOptions{
		Bucket:     b.bucket,
		Object:     oldPath,
		Encryption: encrypt.NewSSE(),
	}
	dstOpts := s3.CopyDestOptions{
		Bucket:     b.bucket,
		Object:     newPath,
		Encryption: encrypt.NewSSE(),
	}

	if _, err = s3Clnt.CopyObject(context.Background(), dstOpts, srcOpts); err != nil {
		return model.NewAppError("moveFile", "api.file.move_file.copy_within_s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err = s3Clnt.RemoveObject(context.Background(), b.bucket, oldPath, s3.RemoveObjectOptions{}); err != nil {
		return model.NewAppError("moveFile", "api.file.move_file.delete_from_s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (b *S3FileBackend) WriteFile(fr io.Reader, path string) (int64, *model.AppError) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return 0, model.NewAppError("WriteFile", "api.file.write_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var contentType string
	path = filepath.Join(b.pathPrefix, path)
	if ext := filepath.Ext(path); model.IsFileExtImage(ext) {
		contentType = model.GetImageMimeType(ext)
	} else {
		contentType = "binary/octet-stream"
	}

	options := s3PutOptions(b.encrypt, contentType)
	info, err := s3Clnt.PutObject(context.Background(), b.bucket, path, fr, -1, options)
	if err != nil {
		return info.Size, model.NewAppError("WriteFile", "api.file.write_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return info.Size, nil
}

func (b *S3FileBackend) AppendFile(fr io.Reader, path string) (int64, *model.AppError) {
	s3Clnt, err := b.s3New()
	if err != nil {
		return 0, model.NewAppError("AppendFile", "api.file.append_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	fp := filepath.Join(b.pathPrefix, path)
	if _, err = s3Clnt.StatObject(context.Background(), b.bucket, fp, s3.StatObjectOptions{}); err != nil {
		return 0, model.NewAppError("AppendFile", "api.file.append_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var contentType string
	if ext := filepath.Ext(fp); model.IsFileExtImage(ext) {
		contentType = model.GetImageMimeType(ext)
	} else {
		contentType = "binary/octet-stream"
	}

	options := s3PutOptions(b.encrypt, contentType)
	sse := options.ServerSideEncryption
	partName := fp + ".part"
	info, err := s3Clnt.PutObject(context.Background(), b.bucket, partName, fr, -1, options)
	defer s3Clnt.RemoveObject(context.Background(), b.bucket, partName, s3.RemoveObjectOptions{})
	if info.Size > 0 {
		src1Opts := s3.CopySrcOptions{
			Bucket: b.bucket,
			Object: fp,
		}
		src2Opts := s3.CopySrcOptions{
			Bucket: b.bucket,
			Object: partName,
		}
		dstOpts := s3.CopyDestOptions{
			Bucket:     b.bucket,
			Object:     fp,
			Encryption: sse,
		}
		_, err = s3Clnt.ComposeObject(context.Background(), dstOpts, src1Opts, src2Opts)
		if err != nil {
			return 0, model.NewAppError("AppendFile", "api.file.append_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
		return info.Size, nil
	}

	var errString string
	if err != nil {
		errString = err.Error()
	}

	return 0, model.NewAppError("AppendFile", "api.file.append_file.s3.app_error", nil, errString, http.StatusInternalServerError)
}

func (b *S3FileBackend) RemoveFile(path string) *model.AppError {
	s3Clnt, err := b.s3New()
	if err != nil {
		return model.NewAppError("RemoveFile", "utils.file.remove_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	if err := s3Clnt.RemoveObject(context.Background(), b.bucket, path, s3.RemoveObjectOptions{}); err != nil {
		return model.NewAppError("RemoveFile", "utils.file.remove_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func getPathsFromObjectInfos(in <-chan s3.ObjectInfo) <-chan s3.ObjectInfo {
	out := make(chan s3.ObjectInfo, 1)

	go func() {
		defer close(out)

		for {
			info, done := <-in

			if !done {
				break
			}

			out <- info
		}
	}()

	return out
}

func (b *S3FileBackend) ListDirectory(path string) (*[]string, *model.AppError) {
	var paths []string

	s3Clnt, err := b.s3New()
	if err != nil {
		return nil, model.NewAppError("ListDirectory", "utils.file.list_directory.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	if !strings.HasSuffix(path, "/") && len(path) > 0 {
		// s3Clnt returns only the path itself when "/" is not present
		// appending "/" to make it consistent across all filesstores
		path = path + "/"
	}

	opts := s3.ListObjectsOptions{
		Prefix: path,
	}
	for object := range s3Clnt.ListObjects(context.Background(), b.bucket, opts) {
		if object.Err != nil {
			return nil, model.NewAppError("ListDirectory", "utils.file.list_directory.s3.app_error", nil, object.Err.Error(), http.StatusInternalServerError)
		}
		paths = append(paths, strings.Trim(object.Key, "/"))
	}

	return &paths, nil
}

func (b *S3FileBackend) RemoveDirectory(path string) *model.AppError {
	s3Clnt, err := b.s3New()
	if err != nil {
		return model.NewAppError("RemoveDirectory", "utils.file.remove_directory.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	opts := s3.ListObjectsOptions{
		Prefix:    filepath.Join(b.pathPrefix, path),
		Recursive: true,
	}
	list := s3Clnt.ListObjects(context.Background(), b.bucket, opts)
	objectsCh := s3Clnt.RemoveObjects(context.Background(), b.bucket, getPathsFromObjectInfos(list), s3.RemoveObjectsOptions{})
	for err := range objectsCh {
		if err.Err != nil {
			return model.NewAppError("RemoveDirectory", "utils.file.remove_directory.s3.app_error", nil, err.Err.Error(), http.StatusInternalServerError)
		}
	}

	return nil
}

func s3PutOptions(encrypted bool, contentType string) s3.PutObjectOptions {
	options := s3.PutObjectOptions{}
	if encrypted {
		options.ServerSideEncryption = encrypt.NewSSE()
	}
	options.ContentType = contentType
	// We set the part size to the minimum allowed value of 5MBs
	// to avoid an excessive allocation in minio.PutObject implementation.
	options.PartSize = 1024 * 1024 * 5

	return options
}

func CheckMandatoryS3Fields(settings *model.FileSettings) *model.AppError {
	if settings.AmazonS3Bucket == nil || len(*settings.AmazonS3Bucket) == 0 {
		return model.NewAppError("S3File", "api.admin.test_s3.missing_s3_bucket", nil, "", http.StatusBadRequest)
	}

	// if S3 endpoint is not set call the set defaults to set that
	if settings.AmazonS3Endpoint == nil || len(*settings.AmazonS3Endpoint) == 0 {
		settings.SetDefaults(true)
	}

	return nil
}
