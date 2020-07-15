// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2015 MASA Group
//
// ****************************************************************************

package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeWriter struct {
	writer io.Writer
}

func (w TimeWriter) Write(p []byte) (int, error) {
	date := time.Now().Format("[2006-01-02 15:04:05] ")
	p = append([]byte(date), p...)
	n, err := w.writer.Write(p)
	if err != nil {
		if n >= len(date) {
			n = n - len(date)
		} else {
			n = 0
		}
	}
	return n, err
}

type RotateWriter struct {
	filename string
	maxFiles int
	maxSize  int64
	size     int64
	inBytes  bool
	file     *os.File
	history  []string
}

func computeMaxSize(maxSize int64, sizeUnit string) int64 {
	if sizeUnit == "kbytes" {
		return maxSize * 1024
	}
	if sizeUnit == "mbytes" {
		return maxSize * 1048576
	}
	return maxSize
}

func (w *RotateWriter) computeSize(info os.FileInfo) int64 {
	if w.inBytes {
		return info.Size()
	}
	// Use an hard-coded line size to prevent time consuming
	// start-ups when re-opening huge log files.
	// see https://masagroup.atlassian.net/browse/SWBUG-14201
	return info.Size() / 200
}

// NewRotateWriter creates a RotateWrite which handles rotating logs.
func NewRotateWriter(filename string, maxFiles int, maxSize int64, sizeUnit string, truncate bool) (*RotateWriter, error) {
	w := &RotateWriter{
		filename: filename,
		maxFiles: maxFiles,
		maxSize:  computeMaxSize(maxSize, sizeUnit),
		inBytes:  sizeUnit == "bytes" || sizeUnit == "kbytes" || sizeUnit == "mbytes",
	}
	err := w.populate()
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(w.filename)
	if err == nil {
		if info.IsDir() {
			return nil, errors.New("invalid filename")
		}
		w.size = w.computeSize(info)
		if truncate {
			err = w.rotate()
			if err != nil {
				return nil, err
			}
		}
	}
	err = w.prune()
	if err != nil {
		return nil, err
	}
	_, err = w.Write(nil)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RotateWriter) Close() error {
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

func (w *RotateWriter) populate() error {
	dir := filepath.Dir(w.filename)
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	stem := strings.TrimSuffix(filepath.Base(w.filename), filepath.Ext(w.filename))
	regex := regexp.MustCompile(`^\Q` + stem + `\E\.\d{8}T\d{6}\.log(\.\d+)*$`)
	for _, info := range entries {
		filename := info.Name()
		if regex.MatchString(filename) {
			w.history = append(w.history, filepath.Join(dir, filename))
		}
	}
	return nil
}

func (w *RotateWriter) rotate() error {
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		if err != nil {
			return err
		}
	}
	w.size = 0
	_, err := os.Stat(w.filename)
	if err == nil {
		filename := appendSuffixToFile(w.filename)
		err := os.Rename(w.filename, filename)
		if err != nil {
			return err
		}
		w.history = append(w.history, filename)
	}
	return nil
}

func appendSuffixToFile(from string) string {
	ext := filepath.Ext(from)
	filename := strings.TrimSuffix(from, ext) +
		"." + time.Now().Format("20060102T150405") + ext
	suffix := 0
	to := filename
	_, err := os.Stat(to)
	for err == nil {
		suffix++
		to = filename + "." + strconv.Itoa(suffix)
		_, err = os.Stat(to)
	}
	return to
}

func (w *RotateWriter) prune() error {
	if w.maxFiles < 0 {
		return nil
	}
	for len(w.history) >= w.maxFiles && len(w.history) > 0 {
		err := os.Remove(w.history[0])
		if err != nil {
			return err
		}
		w.history = w.history[1:]
	}
	return nil
}

func (w *RotateWriter) increaseSize(size int) {
	if w.inBytes {
		w.size += int64(size)
	} else {
		w.size++
	}
}

func lineCount(p []byte) int64 {
	lineSep := []byte{'\n'}
	return int64(bytes.Count(p, lineSep))
}

func (w *RotateWriter) Write(p []byte) (int, error) {
	if w.maxFiles == 0 {
		return len(p), nil
	}
	if w.maxSize > 0 && w.size >= w.maxSize && w.file != nil {
		err := w.rotate()
		if err != nil {
			size, _ := fmt.Fprintf(w.file, "failed to rotate log file: %s\n", err)
			w.increaseSize(size)
		}
		err = w.prune()
		if err != nil {
			size, _ := fmt.Fprintf(w.file, "failed to prune log file: %s\n", err)
			w.increaseSize(size)
		}
	}
	if w.file == nil {
		file, err := os.OpenFile(w.filename, os.O_CREATE+os.O_WRONLY+os.O_APPEND, os.ModePerm)
		if err != nil {
			return 0, err
		}
		w.file = file
	}
	size, err := w.file.Write(p)
	if w.inBytes {
		w.increaseSize(size)
	} else {
		w.size += lineCount(p)
	}
	return size, err
}

type CollapsingWriter struct {
	w     io.Writer
	last  []byte
	count int
}

func (w *CollapsingWriter) Write(p []byte) (int, error) {
	if bytes.Equal(w.last, p) {
		w.count++
		return 0, nil
	}
	if w.count > 0 {
		fmt.Fprintf(w.w, " ...x%d\n", w.count+1)
		w.count = 0
	}
	w.last = make([]byte, len(p))
	copy(w.last, p)
	n, err := w.w.Write(p)
	return n, err
}

func MakeCollapsingWriter(w io.Writer) io.Writer {
	return &CollapsingWriter{w: &TimeWriter{w}}
}
