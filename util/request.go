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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Default http client timeout is no timeout.
var (
	HttpTimeout time.Duration
	Verbose     bool
)

func init() {
	Verbose = os.Getenv("MASA_DEBUG") != ""
}

type HttpError struct {
	message    string
	StatusCode int
}

func (e *HttpError) Error() string {
	return e.message
}

func do(verb, host, path, contentType, SID string, input []byte,
	decode func(http.Header, io.Reader) error) error {

	u := fmt.Sprintf("http://%s%s", host, path)
	if Verbose {
		fmt.Printf("---\n%s %s\n", verb, u)
	}
	rq, err := http.NewRequest(verb, u, bytes.NewBuffer(input))
	if err != nil {
		return err
	}
	if contentType != "" {
		rq.Header.Set("Content-Type", contentType)
	}
	if SID != "" {
		rq.Header.Set("MASA-SID", SID)
	}
	if Verbose {
		for k, values := range rq.Header {
			for _, v := range values {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
		if len(input) > 0 {
			if isBinary(input) {
				fmt.Printf("binary: %d bytes\n", len(input))
			} else {
				fmt.Printf("%v\n", string(input))
			}
		}
	}
	// Disable keep-alive otherwise the client will try and re-use past
	// connections when shutting down and re-starting the server.
	tr := &http.Transport{DisableKeepAlives: true}
	client := http.Client{
		Timeout:   HttpTimeout,
		Transport: tr,
	}
	rsp, err := client.Do(rq)
	if err != nil {
		if Verbose {
			fmt.Printf("error: %s\n\n", err)
		}
		return err
	}
	defer rsp.Body.Close()
	var body io.Reader = rsp.Body
	if Verbose {
		fmt.Println("->")
		for k, values := range rsp.Header {
			for _, v := range values {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
		// This blocks until EOF, which might be different from the actual
		// behaviour of the 'decode' callback provided by the user.
		data, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return err
		}
		fmt.Println(rsp.Status)
		if len(data) > 0 {
			if isBinary(data) {
				fmt.Printf("binary: %d bytes\n", len(data))
			} else {
				fmt.Printf("%v\n", string(data))
			}
		}
		body = bytes.NewBuffer(data)
	}
	if rsp.StatusCode != http.StatusOK &&
		rsp.StatusCode != http.StatusPartialContent {

		data, err := ioutil.ReadAll(body)
		msg := string(data)
		if err != nil {
			msg = err.Error()
		}
		return &HttpError{
			message:    msg,
			StatusCode: rsp.StatusCode,
		}
	}
	if decode == nil {
		return nil
	}
	return decode(rsp.Header, body)
}

func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// Get sends an http GET request and applies a function to the
// response reader.
func Get(host, path, SID string, read func(http.Header, io.Reader) error) error {
	return do("GET", host, path, "application/json", SID, nil, read)
}

// GetJson sends an http GET request and decodes the JSON response into the
// provided output.
func GetJson(host, path, SID string, output interface{}) error {
	return Get(host, path, SID,
		func(_ http.Header, r io.Reader) error {
			return json.NewDecoder(r).Decode(output)
		})
}

// GetString sends an http GET request and returns the response as a string.
func GetString(host, path, SID string) (string, error) {
	data := ""
	err := Get(host, path, SID,
		func(_ http.Header, r io.Reader) error {
			buf := &bytes.Buffer{}
			_, err := io.Copy(buf, r)
			if err != nil {
				return err
			}
			data = buf.String()
			return nil
		})
	return data, err
}

// PostJson sends a JSON http POST request and forwards the response to the
// provided reader.
func Post(host, path, SID string, input interface{}, reader func(http.Header, io.Reader) error) error {
	buf, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}
	return do("POST", host, path, "application/json", SID, buf, reader)
}

// PostJson sends a JSON http POST request and decodes the JSON response into
// the provided output.
func PostJson(host, path, SID string, input, output interface{}) error {
	return Post(host, path, SID, input,
		func(_ http.Header, r io.Reader) error {
			return json.NewDecoder(r).Decode(output)
		})
}

// PostMultipart sends a multipart form file http request reading the file
// content from the provided reader and decodes the JSON response into the
// provided output.
func PostMultipart(host, path, name, SID string, data io.Reader,
	output interface{}) error {

	buffer := &bytes.Buffer{}
	w := multipart.NewWriter(buffer)
	form, err := w.CreateFormFile(name, name)
	if err != nil {
		return err
	}
	_, err = io.Copy(form, data)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return do("POST", host, path, w.FormDataContentType(), SID, buffer.Bytes(),
		func(_ http.Header, r io.Reader) error {
			return json.NewDecoder(r).Decode(output)
		})
}
