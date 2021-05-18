// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils/fileutils"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var publicKey []byte = []byte(`-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2utI7IXzwuDE6no+V8oH
q2sF/NvRRMxgr6a7tlQN2CmYvXW/148BgtISNhgSCKtdFvlaYxAysywJK8LchLIM
v8gs+iCSZfxksg0liq/jqiV/wDW7kJKITsv+T0SFsWWtoEl+4240f5XPySHQphPU
07F/yGzRsjVJEQ1zm4Tlnmh0QqMHYlpGfA9SHMhmmBOn/9BY3YDwoHHtQJ0QdwVN
wFmkqrj4RQy3qdktk3M7kRUI9rwG3E/zGL7JnldbFakS7l8euJ64viuiaTs8nTDz
pDAFNNjvZFGmgfoXJmaW60hbR47IqORzKxI/+njhKg5tQQuVRhtFePyml8ZClKr7
mQIDAQAB
-----END RSA PUBLIC KEY-----`)

//var publicKey []byte = []byte(`-----BEGIN PUBLIC KEY-----
//MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyZmShlU8Z8HdG0IWSZ8r
//tSyzyxrXkJjsFUf0Ke7bm/TLtIggRdqOcUF3XEWqQk5RGD5vuq7Rlg1zZqMEBk8N
//EZeRhkxyaZW8pLjxwuBUOnXfJew31+gsTNdKZzRjrvPumKr3EtkleuoxNdoatu4E
//HrKmR/4Yi71EqAvkhk7ZjQFuF0osSWJMEEGGCSUYQnTEqUzcZSh1BhVpkIkeu8Kk
//1wCtptODixvEujgqVe+SrE3UlZjBmPjC/CL+3cYmufpSNgcEJm2mwsdaXp2OPpfn
//a0v85XL6i9ote2P+fLZ3wX9EoioHzgdgB7arOxY50QRJO7OyCqpKFKv6lRWTXuSt
//hwIDAQAB
//-----END PUBLIC KEY-----`)

func savePKCS8RSAPEMKey(fName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fName)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
	defer outFile.Close()
	//converts a private key to ASN.1 DER encoded form.
	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	err = pem.Encode(outFile, privateKey)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func savePKCS1RSAPublicPEMKey(fName string, pubkey *rsa.PublicKey) {
	//converts an RSA public key to PKCS#1, ASN.1 DER form.
	bytess, err := x509.MarshalPKIXPublicKey(pubkey)
	var pemkey = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytess,
	}
	pemfile, err := os.Create(fName)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
	defer pemfile.Close()
	err = pem.Encode(pemfile, pemkey)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) {
	asn1Bytes, err := asn1.Marshal(pubkey)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	pemfile, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
	defer pemfile.Close()
	err = pem.Encode(pemfile, pemkey)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func SignPKCS1v15(plaintext string, privKey rsa.PrivateKey)  (string) {
	// crypto/rand.Reader is a good source of entropy for blinding the RSA
	// operation.
	rng := rand.Reader
	h := sha512.New()
	h.Write([]byte(plaintext))
	d := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rng, &privKey, crypto.SHA512, d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return "Error from signing"
	}
	return base64.StdEncoding.EncodeToString(signature)
}

func VerifyPKCS1v15(signature string, plaintext string, pubkey rsa.PublicKey) (string) {
	sig, _ := base64.StdEncoding.DecodeString(signature)
	//hashed := sha256.Sum256([]byte(plaintext))
	h := sha512.New()
	h.Write([]byte(plaintext))
	d := h.Sum(nil)
	err := rsa.VerifyPKCS1v15(&pubkey, crypto.SHA512, d, sig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from verification: %s\n", err)
		return "Error from verification:"
	}
	return "Signature Verification Passed"
}

func ValidateLicense(signed []byte) (bool, string) {
	//fmt.Println("signed", signed)
	alicePrivateKey, error := rsa.GenerateKey(rand.Reader, 2048)
	if error != nil {
		fmt.Println(error.Error)
		os.Exit(1)
	}

	// Extract Public Key from RSA Private Key
	alicePublicKey := alicePrivateKey.PublicKey
	secretMessage := `{"id":"w17b7aujupr83ca5te9ba7dygw","issued_at":1553767620980,"starts_at":1553767620980,"expires_at":1556359620980,"sku_name":"Papo - 50 Year End User Subscription","sku_short_name":"","customer":{"id":"s7cawpu9c78tir69jrsitbjdnh","number":88193,"name":"Cong","first_name":"","last_name":"","email":"ductuongquan1@gmail.com","company":"Mr - TRIAL","phone_number":"","address_line1":"","address_line2":"","zip_code":"","city":"","state":"","country":""},"features":{"users":50,"ldap":true,"mfa":true,"google_oauth":true,"office365_oauth":true,"compliance":true,"cluster":true,"custom_brand":true,"mhpns":true,"saml":true,"password_requirements":true,"webrtc":true,"future_features":true,"elastic_search":true}}`
	//fmt.Println("Original Text  ", secretMessage)
	signature1 := SignPKCS1v15(secretMessage, *alicePrivateKey);
	//fmt.Println("Singature :  ", signature1)
	verif := VerifyPKCS1v15(signature1, secretMessage,  alicePublicKey )
	fmt.Println(verif)


	// Method to store the RSA keys in pkcs8 Format
	//savePKCS8RSAPEMKey("/home/cong/go/src/bitbucket.org/enesyteam/papo-server/license/alicepriv.pem", alicePrivateKey)
	// Method to store the RSA keys in pkcs1 Format
	//savePKCS1RSAPublicPEMKey("/home/cong/go/src/bitbucket.org/enesyteam/papo-server/license/alicepub.pem",&alicePublicKey)

	return true, secretMessage


	//decoded := make([]byte, base64.StdEncoding.DecodedLen(len(signed)))
	//
	//_, err := base64.StdEncoding.Decode(decoded, signed)
	//if err != nil {
	//	mlog.Error(fmt.Sprintf("Encountered error decoding license, err=%v", err.Error()))
	//	return false, ""
	//}
	//
	//if len(decoded) <= 256 {
	//	mlog.Error("Signed license not long enough")
	//	return false, ""
	//}
	//
	//// remove null terminator
	//for decoded[len(decoded)-1] == byte(0) {
	//	decoded = decoded[:len(decoded)-1]
	//}
	//
	//plaintext := decoded[:len(decoded)-256]
	//signature := decoded[len(decoded)-256:]
	//
	//
	//
	//block, _ := pem.Decode(publicKey)
	//
	//public, err := x509.ParsePKIXPublicKey(block.Bytes)
	//if err != nil {
	//	mlog.Error(fmt.Sprintf("Encountered error signing license, err=%v", err.Error()))
	//	return false, ""
	//}
	//
	//rsaPublic := public.(*rsa.PublicKey)
	//
	////fmt.Println("rsaPublic", rsaPublic)
	//
	//h := sha512.New()
	//h.Write(plaintext)
	//d := h.Sum(nil)
	//
	//err = rsa.VerifyPKCS1v15(rsaPublic, crypto.SHA512, d, signature)
	//if err != nil {
	//	mlog.Error(fmt.Sprintf("Invalid signature, err=%v", err.Error()))
	//	return false, ""
	//}

	//return true, string(plaintext)
}

func GetAndValidateLicenseFileFromDisk(location string) (*model.License, []byte) {
	fileName := GetLicenseFileLocation(location)

	if _, err := os.Stat(fileName); err != nil {
		mlog.Debug(fmt.Sprintf("We could not find the license key in the database or on disk at %v", fileName))
		return nil, nil
	}

	mlog.Info(fmt.Sprintf("License key has not been uploaded.  Loading license key from disk at %v", fileName))
	licenseBytes := GetLicenseFileFromDisk(fileName)

	if success, licenseStr := ValidateLicense(licenseBytes); !success {
		mlog.Error(fmt.Sprintf("Found license key at %v but it appears to be invalid.", fileName))
		return nil, nil
	} else {
		return model.LicenseFromJson(strings.NewReader(licenseStr)), licenseBytes
	}
}

func GetLicenseFileFromDisk(fileName string) []byte {
	file, err := os.Open(fileName)
	if err != nil {
		mlog.Error(fmt.Sprintf("Failed to open license key from disk at %v err=%v", fileName, err.Error()))
		return nil
	}
	defer file.Close()

	licenseBytes, err := ioutil.ReadAll(file)
	if err != nil {
		mlog.Error(fmt.Sprintf("Failed to read license key from disk at %v err=%v", fileName, err.Error()))
		return nil
	}

	return licenseBytes
}

func GetLicenseFileLocation(fileLocation string) string {
	if fileLocation == "" {
		configDir, _ := fileutils.FindDir("config")
		return filepath.Join(configDir, "mattermost.mattermost-license")
	} else {
		return fileLocation
	}
}

func GetClientLicense(l *model.License) map[string]string {
	props := make(map[string]string)

	props["IsLicensed"] = strconv.FormatBool(l != nil)

	if l != nil {
		props["Id"] = l.Id
		//props["SkuName"] = l.SkuName
		//props["SkuShortName"] = l.SkuShortName
		props["Users"] = strconv.Itoa(*l.Features.Users)
		props["LDAP"] = strconv.FormatBool(*l.Features.LDAP)
		//props["LDAPGroups"] = strconv.FormatBool(*l.Features.LDAPGroups)
		props["MFA"] = strconv.FormatBool(*l.Features.MFA)
		props["SAML"] = strconv.FormatBool(*l.Features.SAML)
		props["Cluster"] = strconv.FormatBool(*l.Features.Cluster)
		props["Metrics"] = strconv.FormatBool(*l.Features.Metrics)
		props["GoogleOAuth"] = strconv.FormatBool(*l.Features.GoogleOAuth)
		props["Office365OAuth"] = strconv.FormatBool(*l.Features.Office365OAuth)
		props["Compliance"] = strconv.FormatBool(*l.Features.Compliance)
		props["MHPNS"] = strconv.FormatBool(*l.Features.MHPNS)
		props["Announcement"] = strconv.FormatBool(*l.Features.Announcement)
		props["Elasticsearch"] = strconv.FormatBool(*l.Features.Elasticsearch)
		props["DataRetention"] = strconv.FormatBool(*l.Features.DataRetention)
		props["IssuedAt"] = strconv.FormatInt(l.IssuedAt, 10)
		props["StartsAt"] = strconv.FormatInt(l.StartsAt, 10)
		props["ExpiresAt"] = strconv.FormatInt(l.ExpiresAt, 10)
		props["Name"] = l.Customer.Name
		props["Email"] = l.Customer.Email
		props["Company"] = l.Customer.Company
		props["PhoneNumber"] = l.Customer.PhoneNumber
		props["EmailNotificationContents"] = strconv.FormatBool(*l.Features.EmailNotificationContents)
		props["MessageExport"] = strconv.FormatBool(*l.Features.MessageExport)
		props["CustomPermissionsSchemes"] = strconv.FormatBool(*l.Features.CustomPermissionsSchemes)
		props["CustomTermsOfService"] = strconv.FormatBool(*l.Features.CustomTermsOfService)
	}

	return props
}
