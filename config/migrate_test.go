// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package config_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"bitbucket.org/enesyteam/papo-server/config"
	"bitbucket.org/enesyteam/papo-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	files := []string{
		"IdpCertificateFile",
		"PublicCertificateFile",
		"PrivateKeyFile",
		"internal.crt",
		"internal2.crt",
	}
	filesData := make([]string, len(files))
	for i := range files {
		// Generate random data for each file, ensuring that stale data from a past test
		// won't generate a false positive.
		filesData[i] = model.NewId()
	}

	setup := func(t *testing.T) {
		os.Clearenv()
		t.Helper()

		tempDir, err := ioutil.TempDir("", "TestMigrate")
		require.NoError(t, err)
		t.Cleanup(func() {
			os.RemoveAll(tempDir)
		})

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		truncateTables(t)
	}

	setupSource := func(t *testing.T, source config.Store) {
		t.Helper()

		cfg := source.Get()
		cfg.ServiceSettings.SiteURL = sToP("http://example.com")
		cfg.SamlSettings.IdpCertificateFile = &files[0]
		cfg.SamlSettings.PublicCertificateFile = &files[1]
		cfg.SamlSettings.PrivateKeyFile = &files[2]
		cfg.PluginSettings.SignaturePublicKeyFiles = []string{
			files[3],
			files[4],
		}
		cfg.SqlSettings.DataSourceReplicas = []string{
			"mysql://mmuser:password@tcp(replicahost:3306)/mattermost",
		}
		cfg.SqlSettings.DataSourceSearchReplicas = []string{
			"mysql://mmuser:password@tcp(searchreplicahost:3306)/mattermost",
		}

		_, err := source.Set(cfg)
		require.NoError(t, err)

		for i, file := range files {
			err = source.SetFile(file, []byte(filesData[i]))
			require.NoError(t, err)
		}
	}

	assertDestination := func(t *testing.T, destination config.Store, source config.Store) {
		t.Helper()

		for i, file := range files {
			hasFile, err := destination.HasFile(file)
			require.NoError(t, err)
			require.Truef(t, hasFile, "destination missing file %s", file)

			actualData, err := destination.GetFile(file)
			require.NoError(t, err)
			assert.Equalf(t, []byte(filesData[i]), actualData, "destination has wrong contents for file %s", file)
		}

		assert.Equal(t, source.Get(), destination.Get())
	}

	t.Run("database to file", func(t *testing.T) {
		setup(t)

		pwd, err := os.Getwd()
		require.NoError(t, err)

		sqlSettings := mainHelper.GetSQLSettings()
		destinationDSN := path.Join(pwd, "config-custom.json")
		sourceDSN := getDsn(*sqlSettings.DriverName, *sqlSettings.DataSource)

		source, err := config.NewDatabaseStore(sourceDSN)
		require.NoError(t, err)
		defer source.Close()

		setupSource(t, source)
		err = config.Migrate(sourceDSN, destinationDSN)
		require.NoError(t, err)

		destination, err := config.NewFileStore(destinationDSN, false)
		require.NoError(t, err)
		defer destination.Close()

		assertDestination(t, destination, source)
	})

	t.Run("file to database", func(t *testing.T) {
		setup(t)

		pwd, err := os.Getwd()
		require.NoError(t, err)

		sqlSettings := mainHelper.GetSQLSettings()
		sourceDSN := path.Join(pwd, "config-custom.json")
		destinationDSN := getDsn(*sqlSettings.DriverName, *sqlSettings.DataSource)

		source, err := config.NewFileStore(sourceDSN, false)
		require.NoError(t, err)
		defer source.Close()

		setupSource(t, source)
		err = config.Migrate(sourceDSN, destinationDSN)
		require.NoError(t, err)

		destination, err := config.NewDatabaseStore(destinationDSN)
		require.NoError(t, err)
		defer destination.Close()

		assertDestination(t, destination, source)
	})
}
