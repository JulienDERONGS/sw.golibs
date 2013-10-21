package restful

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

// Request is a wrapper for a http Request that provides convenience methods
type Request struct {
	Request        *http.Request
	pathParameters map[string]string
}

// PathParameter accesses the Path parameter value by its name
func (self *Request) PathParameter(name string) string {
	return self.pathParameters[name]
}

// QueryParameter returns the (first) Query parameter value by its name
func (self *Request) QueryParameter(name string) string {
	return self.Request.FormValue(name)
}

// ReadEntity check the Accept header and reads the content into the entityReference
func (self *Request) ReadEntity(entityReference interface{}) error {
	defer self.Request.Body.Close()
	buffer, err := ioutil.ReadAll(self.Request.Body)
	if err != nil {
		return err
	}
	field := self.Request.Header.Get(HEADER_ContentType)
	contentType := strings.Split(field, ";")[0]
	contentType = strings.Trim(contentType, " ")
	contentType = strings.ToLower(contentType)
	if contentType == MIME_XML {
		return xml.Unmarshal(buffer, entityReference)
	}
	if contentType == MIME_JSON {
		return json.Unmarshal(buffer, entityReference)
	}
	return errors.New("unknown content-type")
}
