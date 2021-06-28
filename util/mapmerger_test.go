// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2016 MASA Group
//
// ****************************************************************************

package util

import (
	"github.com/masagroup/sw.golibs/swconfig"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestMain(m *testing.M) {
	code := m.Run()
	if code == 0 {
		Cfg.PurgeTestDirs(true)
	}
	DeleteFakeSrcFolder()
	os.Exit(code)
}

var (
	Cfg *swconfig.Config
)

func init() {
	CreateFakeSrcFolder()
	Cfg = swconfig.ParseFlags()
}

func TestMergeMap(t *testing.T) {
	first := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"key3": "value3",
			"key4": map[string]interface{}{
				"key5": "value5",
				"key6": "value6",
			},
		},
	}
	second := map[string]interface{}{
		"key7": "value7",
		"key8": "value8",
		"key2": map[string]interface{}{
			"key9": "value9",
			"key4": map[string]interface{}{
				"key10": "value10",
			},
		},
	}
	result, err := MergeMap(first, second)
	assert.NoError(t, err)
	assert.Equal(t, result, map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"key3": "value3",
			"key9": "value9",
			"key4": map[string]interface{}{
				"key5":  "value5",
				"key6":  "value6",
				"key10": "value10",
			},
		},
		"key7": "value7",
		"key8": "value8",
	})
}

func CreateFakeSrcFolder() {
	os.MkdirAll(filepath.Join(".", "src"), os.ModePerm)
}
 
func DeleteFakeSrcFolder() {
	os.RemoveAll("./src/")
}
