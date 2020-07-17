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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func loadJson(path string, data interface{}) error {
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

func saveJson(path string, data interface{}) error {
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

type Config struct {
	flags    map[string]string
	excludes map[string]struct{}
	path     string
}

func (c *Config) parseFlags(path string) error {
	c.path = path
	c.flags = make(map[string]string)
	if path != "" {
		err := loadJson(c.path, &c.flags)
		if err != nil {
			return err
		}
	}
	flag.Visit(func(flag *flag.Flag) {
		_, ok := c.flags[flag.Name]
		if ok {
			delete(c.flags, flag.Name)
		}
	})
	flag.VisitAll(func(flag *flag.Flag) {
		val, ok := c.flags[flag.Name]
		if ok {
			_ = flag.Value.Set(val)
		}
	})
	return nil
}

func (c *Config) saveFlags() error {
	flag.VisitAll(func(flag *flag.Flag) {
		name := flag.Name
		if _, ok := c.excludes[name]; !ok {
			c.flags[name] = flag.Value.String()
		}
	})
	if c.path != "" {
		return saveJson(c.path, &c.flags)
	}
	return nil
}

func NewConfig(path string, excludes []string) (*Config, error) {
	ignores := map[string]struct{}{}
	for _, v := range excludes {
		ignores[v] = struct{}{}
	}
	c := &Config{
		excludes: ignores,
	}
	return c, c.Parse(path)
}

// Parse parses the given configuration file and updates it with the
// currently defined flags.
// The file gets created if it doesn't exist.
func (c *Config) Parse(path string) error {
	err := c.parseFlags(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to load config file: %v", err)
	}
	return c.saveFlags()
}

func (c *Config) Update(key, value string) error {
	c.flags[key] = value
	if c.path != "" {
		return saveJson(c.path, &c.flags)
	}
	return nil
}

func (c *Config) GetFlag(key string) string {
	value, ok := c.flags[key]
	if ok {
		return value
	}
	return ""
}
