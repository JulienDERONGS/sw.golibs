// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2016 MASA Group
//
// ****************************************************************************
package swconfig

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Platform     string
	BaseDir      string
	BinDir       string
	DataDir      string
	OutDir       string
	RunDir       string
	ExercisesDir string
	ShowLog      bool
	KeepOutput   bool
	Debug        bool
	Checkpoint   bool
	RedundantMsg bool
	Repeater     bool
	Gaming       bool
	TestPort     int
	Timeout      time.Duration
	dirs         []string
	DumpLevel    string
}

func getBaseDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("cannot locate base dir: %s", err))
	}
	for root := cwd; root != ""; root = filepath.Dir(root) {
		_, err = os.Stat(filepath.Join(root, "src"))
		if err == nil {
			return root
		}
	}
	return ""
}

func ParseFlags() *Config {
	cfg := Config{}
	flag.StringVar(&cfg.Platform, "platform", "",
		"platform for which to run the tests")
	flag.StringVar(&cfg.BinDir, "bin-dir", "",
		"path to bin directory, defaults to 'root-dir'/bin")
	flag.StringVar(&cfg.DataDir, "data-dir", "",
		"path to data directory, defaults to 'root-dir'/data")
	flag.StringVar(&cfg.RunDir, "run-dir", "",
		"path application run directory, defaults to 'root-dir'/run/'platform'")
	flag.StringVar(&cfg.OutDir, "out-dir", "",
		"path to applications directory, defaults to 'root-dir'/out/'platform'/release")
	flag.StringVar(&cfg.ExercisesDir, "exercises-dir", "",
		"exercises directory, default to 'root-dir'/data/tests/gosword")
	flag.IntVar(&cfg.TestPort, "test-port", 35000,
		"base port for spawned simulations")
	flag.BoolVar(&cfg.KeepOutput, "keep-output", false,
		"keep temporary output folders")
	flag.BoolVar(&cfg.Checkpoint, "checkpoint", false,
		"create a checkpoint, reload it and compare data when closing simulation")
	flag.BoolVar(&cfg.RedundantMsg, "redundant-msg", false,
		"check redundant msg between 2 ticks")
	flag.BoolVar(&cfg.Repeater, "repeater", false,
		"create a repeater, connect it to the simulation and use it for client connections")
	flag.BoolVar(&cfg.Gaming, "gaming", false,
		"start gaming after starting the simulation and before connecting the test client")
	flag.BoolVar(&cfg.ShowLog, "show-log", false, "print simulation log files")
	flag.BoolVar(&cfg.Debug, "d", false, "use debug executables")
	flag.StringVar(&cfg.DumpLevel, "dump-level", "normal", "dump level normal/full/off")
	flag.Parse()
	if cfg.Platform == "" {
		// Set default platform
		cfg.Platform = "vc140_x64"
	}
	cfg.BaseDir = getBaseDir()
	if cfg.BinDir == "" {
		cfg.BinDir = filepath.Join(cfg.BaseDir, "bin")
	}
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(cfg.BaseDir, "data")
	}
	if cfg.RunDir == "" {
		cfg.RunDir = filepath.Join(cfg.BaseDir, "run", cfg.Platform)
	}
	if cfg.DumpLevel == "" {
		cfg.DumpLevel = "normal"
	}
	if cfg.OutDir == "" {
		if cfg.Debug {
			cfg.OutDir = filepath.Join(cfg.BaseDir, "out", cfg.Platform, "debug")
		} else {
			cfg.OutDir = filepath.Join(cfg.BaseDir, "out", cfg.Platform, "release")
		}
	}
	if cfg.ExercisesDir == "" {
		cfg.ExercisesDir = filepath.Join(cfg.BaseDir, "data/tests/gosword")
	}
	cfg.Timeout = 2 * time.Minute
	timeoutStr := os.Getenv("MASA_TIMEOUT")
	if timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err == nil {
			cfg.Timeout = t
		}
	}
	return &cfg
}

// MakeTestDir creates a uniquely named test directory inside
// 'project root'/out/prefix and adds it to the internal list of folders.
// All folders can be purged with PurgeTestDirs.
func (c *Config) MakeTestDir(prefix string) (string, error) {
	baseDir := filepath.Join(c.BaseDir, "out", prefix)
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return "", err
	}
	dir, err := ioutil.TempDir(baseDir, time.Now().Format("20060102T150405-"))
	if err == nil {
		c.AddTestDir(dir)
	}
	return dir, err
}

// AddTestDir registers the given folder to be deleted.
// All folders can be purged with PurgeTestDirs.
func (c *Config) AddTestDir(dir string) {
	c.dirs = append(c.dirs, dir)
}

// PurgeTestDirs clears the folder list optionally removing directories
func (c *Config) PurgeTestDirs(remove bool) {
	dirs := c.dirs
	c.dirs = []string{}
	if !c.KeepOutput && remove {
		for _, dir := range dirs {
			// Mitigate the infamous windows folder tree deletion bug
			err := os.RemoveAll(dir)
			if err != nil {
				err = os.RemoveAll(dir)
				if err != nil {
					fmt.Printf("cannot clean up test folder %s: %s\n", dir, err)
				}
			}
		}
	}
}

func (cfg *Config) FindApplication(name string) string {
	app := filepath.Join(cfg.BinDir, name+".exe")
	_, err := os.Stat(app)
	if err == nil {
		return app
	}
	if cfg.Debug {
		return filepath.Join(cfg.OutDir, name+"_d.exe")
	}
	return filepath.Join(cfg.OutDir, name+".exe")
}
