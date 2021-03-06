// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bitbucket.org/enesyteam/papo-server/mlog"
	"bitbucket.org/enesyteam/papo-server/model"
	"bitbucket.org/enesyteam/papo-server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupervisor(t *testing.T) {
	for name, f := range map[string]func(*testing.T){
		"Supervisor_InvalidExecutablePath":     testSupervisorInvalidExecutablePath,
		"Supervisor_NonExistentExecutablePath": testSupervisorNonExistentExecutablePath,
		"Supervisor_StartTimeout":              testSupervisorStartTimeout,
	} {
		t.Run(name, f)
	}
}

func testSupervisorInvalidExecutablePath(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	ioutil.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{"id": "foo", "backend": {"executable": "/foo/../../backend.exe"}}`), 0600)

	bundle := model.BundleInfoForPath(dir)
	log := mlog.NewLogger(&mlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "error",
		EnableFile:    false,
	})
	supervisor, err := newSupervisor(bundle, nil, log, nil)
	assert.Nil(t, supervisor)
	assert.Error(t, err)
}

func testSupervisorNonExistentExecutablePath(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	ioutil.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{"id": "foo", "backend": {"executable": "thisfileshouldnotexist"}}`), 0600)

	bundle := model.BundleInfoForPath(dir)
	log := mlog.NewLogger(&mlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "error",
		EnableFile:    false,
	})
	supervisor, err := newSupervisor(bundle, nil, log, nil)
	require.Error(t, err)
	require.Nil(t, supervisor)
}

// If plugin development goes really wrong, let's make sure plugin activation won't block forever.
func testSupervisorStartTimeout(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	backend := filepath.Join(dir, "backend.exe")
	utils.CompileGo(t, `
		package main

		func main() {
			for {
			}
		}
	`, backend)

	ioutil.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{"id": "foo", "backend": {"executable": "backend.exe"}}`), 0600)

	bundle := model.BundleInfoForPath(dir)
	log := mlog.NewLogger(&mlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "error",
		EnableFile:    false,
	})
	supervisor, err := newSupervisor(bundle, nil, log, nil)
	require.Error(t, err)
	require.Nil(t, supervisor)
}
