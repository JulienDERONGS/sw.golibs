// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2016 MASA Group
//
// ****************************************************************************

package log

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	content  = "some content"
	someline = "some line\n"
	filename = "filename.log"
	created  = `^filename\.\d{8}T\d{6}\.log.*$`
)

func makeDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "writers-test")
	assert.NoError(t, err)
	return dir
}

func checkWrite(t *testing.T, w io.Writer) {
	i, err := w.Write([]byte(content))
	assert.NoError(t, err)
	assert.Equal(t, i, len(content))
}

func checkWriteLine(t *testing.T, w io.Writer) {
	i, err := w.Write([]byte(someline))
	assert.NoError(t, err)
	assert.Equal(t, i, len(someline))
}

func checkContent(t *testing.T, filename, text string) {
	b, err := ioutil.ReadFile(filename)
	assert.NoError(t, err)
	assert.Equal(t, string(b), text)
}

func readFiles(t *testing.T, dir string) []string {
	entries, err := ioutil.ReadDir(dir)
	assert.NoError(t, err)
	files := []string{}
	for _, info := range entries {
		files = append(files, info.Name())
	}
	return files
}

func TestRotatingLogInvalidFilenameReturnsErrorUponCreation(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, ".."), 0, 3, "bytes", true)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func TestRotatingLogMaxFilesSetToZeroDisablesLog(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, filename), 0, 3, "bytes", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	assert.Len(t, readFiles(t, dir), 0)
}

func TestRotatingLogMaxSizeSetToZeroDisablesRotation(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, filename), 2, 0, "bytes", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	checkWrite(t, w)
	files := readFiles(t, dir)
	assert.Len(t, files, 1)
	assert.Equal(t, files[0], filename)
	checkContent(t, filepath.Join(dir, files[0]), content+content)
}

func TestRotatingLogSwitchesToNextLogAfterMaxFilesIsReached(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, filename), 2, int64(len(content)), "bytes", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	checkWrite(t, w)
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	checkContent(t, filepath.Join(dir, files[1]), content)
	checkContent(t, filepath.Join(dir, files[0]), content)
}

func TestRotatingLogWithNegativeMaxFiles(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, filename), -1, int64(len(content)), "bytes", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	checkWrite(t, w)
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	checkContent(t, filepath.Join(dir, files[1]), content)
	checkContent(t, filepath.Join(dir, files[0]), content)
}

func TestRotatingLogNonInitiallyEmptySwitchesToNextLogAfterMaxSizeIsReached(t *testing.T) {
	dir := makeDir(t)
	log := filepath.Join(dir, filename)
	initial := "initial"
	err := ioutil.WriteFile(log, []byte(initial), os.ModePerm)
	assert.NoError(t, err)
	w, err := NewRotateWriter(log, 2, int64(len(content)+len(initial)), "bytes", false)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	checkWrite(t, w)
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	checkContent(t, filepath.Join(dir, files[1]), content)
	checkContent(t, filepath.Join(dir, files[0]), "initial"+content)
}

func TestRotatingLogDeletesOldestLogWhenMaxFilesIsReached(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, filename), 2, int64(len(content)), "bytes", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	checkWrite(t, w)
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	file := files[0]
	checkWrite(t, w)
	files = readFiles(t, dir)
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	assert.NotEqual(t, files[0], file)
}

var (
	rotated1   = "filename.20130916T115500.log"
	rotated2   = "filename.20130916T115501.log"
	unrelated1 = "filename.20130916T115503.log_unrelated"
	unrelated2 = "unrelated_1.20130916T115503.log"
	unrelated3 = "unrelated_filename.20130916T115503.log"
)

func makeFiles(t *testing.T) string {
	dir := makeDir(t)
	file, err := os.Create(filepath.Join(dir, rotated1))
	assert.NoError(t, err)
	file.Close()
	file, err = os.Create(filepath.Join(dir, rotated2))
	assert.NoError(t, err)
	file.Close()
	file, err = os.Create(filepath.Join(dir, unrelated1))
	assert.NoError(t, err)
	file.Close()
	file, err = os.Create(filepath.Join(dir, unrelated2))
	assert.NoError(t, err)
	file.Close()
	file, err = os.Create(filepath.Join(dir, unrelated3))
	assert.NoError(t, err)
	file.Close()
	return dir
}

func testRelisting(t *testing.T, initial string, maxFiles int, maxSize int64, truncate bool, expected ...string) {
	dir := makeFiles(t)
	defer os.RemoveAll(dir)
	if initial != "" {
		err := ioutil.WriteFile(filepath.Join(dir, filename), []byte(initial), os.ModePerm)
		assert.NoError(t, err)
	}
	w, err := NewRotateWriter(filepath.Join(dir, filename), maxFiles, maxSize, "bytes", truncate)
	assert.NoError(t, err)
	defer w.Close()
	actual := readFiles(t, dir)
	assert.Len(t, actual, len(expected))
	for i, e := range expected {
		if e == created {
			assert.NotRegexp(t, actual[i], e)
			assert.NotEqual(t, actual[i], rotated1)
			assert.NotEqual(t, actual[i], rotated2)
			assert.NotEqual(t, actual[i], filename)
		} else {
			assert.Equal(t, actual[i], e)
		}
	}
}

func TestRotatingLogRelistsExistingRotatedLogFiles(t *testing.T) {
	testRelisting(t, content, 3, 3, true, rotated2, unrelated1, created, filename, unrelated2, unrelated3)
	testRelisting(t, content, 3, 3, false, rotated1, rotated2, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, content, 2, 3, true, unrelated1, created, filename, unrelated2, unrelated3)
	testRelisting(t, content, 2, 3, false, rotated2, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, content, 1, 3, true, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, content, 1, 3, false, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, content, 0, 3, true, unrelated1, unrelated2, unrelated3)
	testRelisting(t, content, 0, 3, false, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 3, 3, true, rotated1, rotated2, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 3, 3, false, rotated1, rotated2, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 2, 3, true, rotated2, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 2, 3, false, rotated2, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 1, 3, true, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 1, 3, false, unrelated1, filename, unrelated2, unrelated3)
	testRelisting(t, "", 0, 3, true, unrelated1, unrelated2, unrelated3)
	testRelisting(t, "", 0, 3, false, unrelated1, unrelated2, unrelated3)
}

func TestRotatingLogSupportsLinesAsSizeUnit(t *testing.T) {
	dir := makeDir(t)
	w, err := NewRotateWriter(filepath.Join(dir, filename), 5, 2, "lines", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWriteLine(t, w)
	checkWriteLine(t, w)
	checkWriteLine(t, w)
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	checkContent(t, filepath.Join(dir, files[0]), someline+someline)
	checkContent(t, filepath.Join(dir, files[1]), someline)
}

func TestRotatingLogSupportsManyLinesAsSizeUnit(t *testing.T) {
	dir := makeDir(t)
	nLines := 7
	w, err := NewRotateWriter(filepath.Join(dir, filename), 5, int64(nLines), "lines", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWriteLine(t, w)
	for i := 0; i < nLines; i++ {
		checkWriteLine(t, w)
	}
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	checkContent(t, filepath.Join(dir, files[0]), strings.Repeat(someline, nLines))
	checkContent(t, filepath.Join(dir, files[1]), someline)
}

func TestRotatingLogCountsLinesWell(t *testing.T) {
	dir := makeDir(t)
	nLines := 3
	w, err := NewRotateWriter(filepath.Join(dir, filename), 5, int64(nLines), "lines", true)
	assert.NoError(t, err)
	defer w.Close()
	checkWrite(t, w)
	checkWriteLine(t, w)
	for i := 0; i < nLines; i++ {
		checkWriteLine(t, w)
	}
	files := readFiles(t, dir)
	assert.Len(t, files, 2)
	assert.Equal(t, files[1], filename)
	assert.NotRegexp(t, files[0], created)
	checkContent(t, filepath.Join(dir, files[0]), content+strings.Repeat(someline, nLines))
	checkContent(t, filepath.Join(dir, files[1]), someline)
}

func TestCollapsingLog(t *testing.T) {
	b := bytes.Buffer{}
	w := CollapsingWriter{w: &b}
	m := []byte("message")
	n, err := w.Write(m)
	assert.NoError(t, err)
	assert.Equal(t, n, len(m))
	assert.Equal(t, b.String(), "message")
	b.Reset()
	n, err = w.Write(m)
	assert.NoError(t, err)
	assert.Equal(t, n, 0)
	assert.Equal(t, b.String(), "")
	n, err = w.Write(m)
	assert.NoError(t, err)
	assert.Equal(t, n, 0)
	assert.Equal(t, b.String(), "")
	m2 := []byte("another message")
	n, err = w.Write(m2)
	assert.NoError(t, err)
	assert.Equal(t, n, len(m2))
	assert.Equal(t, b.String(), " ...x3\nanother message")
	b.Reset()
	n, err = w.Write(m)
	assert.NoError(t, err)
	assert.Equal(t, n, len(m))
	assert.Equal(t, b.String(), "message")
}
