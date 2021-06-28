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
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func LoadJson(path string, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func SaveJson(path string, data interface{}) error {
	buf, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir+0755)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, bytes.NewReader(buf))
	return err
}
