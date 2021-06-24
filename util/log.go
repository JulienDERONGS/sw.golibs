// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2015 MASA Group
//
// ****************************************************************************

package util

import (
	"flag"
	"github.com/go-errors/errors"
	"github.com/masagroup/sw.golibs/config"
	masalog "github.com/masagroup/sw.golibs/log"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// CatchPanic recovers from any panic and logs the stack trace.
func CatchPanic() {
	if r := recover(); r != nil {
		log.Printf("panic: %s\n", errors.Wrap(r, 2).ErrorStack())
	}
}

// LogPanic recovers from any panic, logs the stack trace, and re-panics.
func LogPanic() {
	if r := recover(); r != nil {
		log.Printf("panic: %s\n", errors.Wrap(r, 2).ErrorStack())
		panic(r)
	}
}

// Go starts a goroutine on the supplied function wrapping it with a recover to
// log any panic which may occur.
func Go(f func()) {
	go func() {
		defer LogPanic()
		f()
	}()
}

func makeDebugLog(path string) (string, *os.File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	err = os.MkdirAll(filepath.Dir(abs), os.ModePerm)
	if err != nil {
		return "", nil, err
	}
	f, err := os.OpenFile(abs, os.O_WRONLY+os.O_APPEND+os.O_CREATE, os.ModePerm)
	return abs, f, err
}

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
)

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

// redirectStderr to the file passed in
func redirectStderr(f *os.File) {
	err := setStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
	if err != nil {
		log.Printf("Failed to redirect stderr to file: %s", err)
	}
}

// Parse configures the default logger and parses application arguments.
// It returns a closer and the configuration file path.
// Stdout will receive all logs.
// Start-up logs are sent to Debug/'logPrefix'.log and can be used to diagnose
// flag parse failures.
// If a log file is configured (using the -log flag) all logs will be sent to
// that file except for stderr which goes to a generic debug file.
func Parse(logPrefix string) (io.Closer, *config.Config) {
	log.SetOutput(masalog.MakeCollapsingWriter(os.Stderr))
	log.SetPrefix("<" + logPrefix + "> ")
	log.SetFlags(0)
	debug, f, err := makeDebugLog(filepath.Join(filepath.Dir(os.Args[0]), "Debug", logPrefix+".log"))
	if err == nil {
		defer f.Close()
		log.SetOutput(masalog.MakeCollapsingWriter(io.MultiWriter(f, os.Stdout)))
		redirectStderr(f)
	}
	log.Println("command line", os.Args)
	file := flag.String("log", "", "optional log filename")
	maxFiles := flag.Int("max-files", -1, "number of log files to keep when rotating, a negative value means infinite, defaults to -1")
	maxSize := flag.Int64("max-size", 100, "log size in bytes to reach before rotating, defaults to 100, 0 disables rotation")
	sizeUnit := flag.String("size-unit", "mbytes", "log size unit, valid values are 'bytes', 'kbytes', 'mbytes' or 'lines'")
	debugPort := flag.Int("debug-port", 0, "start pprof http debug server on supplied port number")
	configFile := flag.String("config", "", "config filename")
	flag.Parse()
	config, err := config.NewConfig(*configFile,
		[]string{
			"config",
			"register",
			"unregister",
			"daemon",
		})
	if err != nil {
		log.Fatalf("unable to parse flags : %v", err)
	}
	wd, err := os.Getwd()
	if err == nil {
		log.Println("working-directory", wd)
	} else {
		log.Println("working-directory", err)
	}
	var c io.Closer
	if len(*file) > 0 && *maxFiles != 0 {
		dir := filepath.Dir(*file)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("unable to create log file directory %v: %v", dir, err)
		}
		w, err := masalog.NewRotateWriter(*file, *maxFiles, *maxSize, *sizeUnit, true)
		if err != nil {
			log.Fatalf("unable to create log file %v: %v", *file, err)
		}
		log.SetOutput(masalog.MakeCollapsingWriter(io.MultiWriter(w, os.Stdout)))
		c = w
	} else {
		log.SetOutput(masalog.MakeCollapsingWriter(os.Stdout))
	}
	log.Println("Sword " + logPrefix + " " + SWORD_VERSION + " - copyright Masa Group 2016")
	log.Println("command line", os.Args)
	log.Println("debug", debug)
	if len(*configFile) > 0 {
		log.Println("config", *configFile)
	}
	if len(*file) > 0 {
		log.Println("log", *file)
		if *maxFiles < 0 {
			log.Println("max-files", *maxFiles, "(infinite)")
		} else {
			log.Println("max-files", *maxFiles)
		}
		log.Println("max-size", *maxSize)
		log.Println("size-unit", *sizeUnit)
	}
	if *debugPort > 0 {
		log.Println("debug-port", *debugPort)
		Go(func() {
			log.Println(http.ListenAndServe(":"+strconv.Itoa(*debugPort), nil))
		})
	}
	return c, config
}
