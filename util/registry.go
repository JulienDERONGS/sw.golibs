// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2014 MASA Group
//
// ****************************************************************************

// +build windows

package util

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	regRoot = syscall.HKEY_CURRENT_USER
	regPath = `SOFTWARE\Masa GROUP\Sword\`
)

var (
	advapi32         = syscall.NewLazyDLL("advapi32.dll")
	procRegCloseKey  = advapi32.NewProc("RegCloseKey")
	procRegCreateKey = advapi32.NewProc("RegCreateKeyExW")
	procRegOpenKey   = advapi32.NewProc("RegOpenKeyExW")
)

func regCloseKey(hkey uint) error {
	ret, _, _ := procRegCloseKey.Call(uintptr(hkey))
	if ret != 0 {
		return fmt.Errorf("unable to close registry key %v", hkey)
	}
	return nil
}

func RegOpenKey(path string) bool {
	var result uint
	ret, _, _ := procRegOpenKey.Call(
		regRoot,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(regPath+path))),
		0,
		uintptr(syscall.KEY_READ),
		uintptr(unsafe.Pointer(&result)),
	)
	if ret != 0 {
		return false
	}
	regCloseKey(result)
	return true
}

func RegCreateKey(path string) error {
	var result uint
	ret, _, _ := procRegCreateKey.Call(
		uintptr(regRoot),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(regPath+path))),
		0, 0, 0,
		uintptr(syscall.KEY_ALL_ACCESS),
		0,
		uintptr(unsafe.Pointer(&result)),
		0,
	)
	if ret != 0 {
		return fmt.Errorf("unable to load registry key %s", path)
	}
	return regCloseKey(result)
}

func RegQueryValueString(path, name string) (string, error) {
	var key syscall.Handle
	err := syscall.RegOpenKeyEx(
		regRoot,
		syscall.StringToUTF16Ptr(regPath+path),
		0,
		syscall.KEY_READ,
		&key)
	if err != nil {
		return "", err
	}
	defer syscall.RegCloseKey(key)

	var typ uint32
	var buf [512]uint16
	n := uint32(len(buf) * 2)
	if syscall.RegQueryValueEx(key,
		syscall.StringToUTF16Ptr(name),
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&buf[0])),
		&n) != nil {
		return "", fmt.Errorf("unable to load query value %s", name)
	}
	return syscall.UTF16ToString(buf[:]), nil
}
