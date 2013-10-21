package restful

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestPathParameter(t *testing.T) {
	hreq := http.Request{Method: "GET"}
	hreq.URL, _ = url.Parse("http://www.google.com/search?q=foo&q=bar")
	rreq := Request{http.Request: &hreq}
	if rreq.QueryParameter("q") != "foo" {
		t.Errorf("q!=foo %#v", rreq)
	}
}

type Message struct {
	Name string
}

func TestAcceptHeader(t *testing.T) {
	hreq := http.Request{Method: "GET"}
	hreq.Header = http.Header{}
	hreq.Header.Set("Content-Type", "application/JSON; charset=UTF-8")
	hreq.Body = ioutil.NopCloser(bytes.NewBufferString("{\"name\":\"foo\"}"))
	rreq := Request{http.Request: &hreq}
	msg := Message{}
	err := rreq.ReadEntity(&msg)
	if err != nil {
		t.Errorf("err!=nil %#v", err)
	}
	if msg.Name != "foo" {
		t.Errorf("output!=foo %#v", msg.Name)
	}
}
