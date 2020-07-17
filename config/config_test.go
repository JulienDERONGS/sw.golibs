// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2016 MASA Group
//
// ****************************************************************************

package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func makeDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "config-test")
	assert.NoError(t, err)
	return dir
}

func TestParseConfig(t *testing.T) {
	flag.String("config-test", "", "config test flag")
	flag.String("ignore", "", "flag ignored")

	dir := makeDir(t)
	configFile := filepath.Join(dir, "config.cfg")
	config, err := NewConfig(configFile, []string{"ignore"})
	assert.NoError(t, err)

	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	testFlag := flag.Lookup("config-test")
	assert.Equal(t, testFlag.Value.String(), "")

	err = config.Update("config-test", "test")
	assert.NoError(t, err)
	err = config.Parse(configFile)
	assert.NoError(t, err)

	assert.Equal(t, config.GetFlag("config-test"), "test")

	testFlag = flag.Lookup("config-test")
	assert.Equal(t, testFlag.Value.String(), "test")
}
