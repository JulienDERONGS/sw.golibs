// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2014 MASA Group
//
// ****************************************************************************
package windows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var (
	rePlatform = regexp.MustCompile(`(\d\d)\.\d\d\.\d+.*(x64|x86)`)
)

// Returns poney Visual Studio architecture from cl.exe arch string.
func getArchSuffix(s string) (string, error) {
	switch s {
	case "x64":
		return "_x64", nil
	case "x86":
		return "", nil
	}
	return "", fmt.Errorf("unknown architecture identifier: %s", s)
}

// Returns poney Visual Studio identifier from its cl.exe version
func getMsvcVersion(s string) (string, error) {
	switch s {
	case "14":
		return "vc80", nil
	case "16":
		return "vc100", nil
	case "19":
		return "vc140", nil
	}
	return "", fmt.Errorf("unknown MSVC version: %s, you need to update getMsvcVersion in mbuild.go", s)
}

// Returns poney platform string from the current environment, or an error.
func GetPlatform() (string, error) {
	cmd := exec.Command("cl.exe")
	data, err := cmd.CombinedOutput()
	output := string(data)
	if err != nil {
		return "", err
	}
	m := rePlatform.FindStringSubmatch(output)
	if m == nil {
		return "", fmt.Errorf("could not extract platform from:\n%s", output)
	}
	version, err := getMsvcVersion(m[1])
	if err != nil {
		return "", err
	}
	arch, err := getArchSuffix(m[2])
	if err != nil {
		return "", err
	}
	return version + arch, nil
}

// GetFromDir returns the directory from which the given file is found, starting
// from the given directory and moving up the parent directories one by one.
func GetFromDir(file, dir string) (string, error) {
	path := dir
	for {
		_, err := os.Stat(filepath.Join(path, file))
		if err == nil {
			return path, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("cannot find '%s' from: %s", file, dir)
		}
		path = parent
	}
}

// GetRootDir returns the root directory of the enclosing project.
func GetRootDir() (string, error) {
	origin, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not retrieve working directory: %s", err)
	}
	return GetFromDir("build/build.xml", origin)
}
