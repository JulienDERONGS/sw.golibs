// ****************************************************************************
//
// This file is part of a MASA library or program.
// Refer to the included end-user license agreement for restrictions.
//
// Copyright (c) 2013 MASA Group
//
// ****************************************************************************
package ts

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Translation struct {
	Type string `xml:"type,attr,omitempty"`
	Text string `xml:",innerxml"`
}

type Source struct {
	Text string `xml:",innerxml"`
}

type TranslatorComment struct {
	Text string `xml:",innerxml"`
}

type OldSource struct {
	Text string `xml:",innerxml"`
}

type Message struct {
	Numerus           string             `xml:"numerus,attr,omitempty"`
	Source            *Source            `xml:"source"`
	TranslatorComment *TranslatorComment `xml:"translatorcomment,omitempty"`
	Translation       *Translation       `xml:"translation"`
	Id                string             `xml:"id,attr,omitempty"`
	OldSource         *OldSource         `xml:"oldsource,omitempty"`
	Utf8              bool               `xml:"utf8,attr,omitempty"`
}

type Messages []*Message

func (m Messages) Len() int {
	return len(m)
}

func (m Messages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Messages) Less(i, j int) bool {
	return m[i].Source.Text < m[j].Source.Text
}

type Context struct {
	Name     string   `xml:"name"`
	Messages Messages `xml:"message"`
}

func (c Context) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}
	e.Indent("", "    ")
	err = e.EncodeElement(c.Name, xml.StartElement{Name: xml.Name{Local: "name"}})
	if err != nil {
		return err
	}
	err = e.EncodeElement(c.Messages, xml.StartElement{Name: xml.Name{Local: "message"}})
	if err != nil {
		return err
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

type Contexts []*Context

type TS struct {
	Lang           string   `xml:"language,attr"`
	SourceLanguage string   `xml:"sourcelanguage,attr"`
	Version        string   `xml:"version,attr"`
	Contexts       Contexts `xml:"context"`
	Path           string   `xml:"-" json:"-"`
}

func (ts *TS) Format() (string, error) {
	output, err := xml.Marshal(*ts)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	_, err = buffer.WriteString(strings.ToLower(xml.Header) + "<!DOCTYPE TS>\n")
	if err != nil {
		return "", err
	}
	_, err = buffer.Write(output)
	if err != nil {
		return "", err
	}
	_, err = buffer.WriteString("\n")
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (ts *TS) Write() error {
	value, err := ts.Format()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ts.Path, []byte(value), 0700)
}

func (ts TS) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "version"}, Value: ts.Version})
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "language"}, Value: ts.Lang})
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "sourcelanguage"}, Value: ts.SourceLanguage})
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}
	e.Indent("\n", "    ")
	err = e.EncodeElement(ts.Contexts, xml.StartElement{Name: xml.Name{Local: "context"}})
	if err != nil {
		return err
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func (ts *TS) FindTranslation(context string, source string) *Translation {
	for _, ctx := range ts.Contexts {
		if len(context) != 0 && ctx.Name != context {
			continue
		}
		for _, message := range ctx.Messages {
			if html.UnescapeString(message.Source.Text) == html.UnescapeString(source) {
				return message.Translation
			}
		}
	}
	return nil
}

type Translations map[string]*TS

func (ts *Translations) Apply(f func(lang, context string, msg *Message)) {
	for lang, translation := range *ts {
		for _, context := range translation.Contexts {
			for _, message := range context.Messages {
				f(lang, context.Name, message)
			}
		}
	}
}

func (ts *Translations) Write() error {
	for _, translation := range *ts {
		err := translation.Write()
		if err != nil {
			return err
		}
	}
	return nil
}

func readTranslationData(data []byte) (*TS, error) {
	x := &TS{}
	err := xml.Unmarshal(data, x)
	if err != nil {
		return nil, fmt.Errorf("cannot parse ts: %v", err)
	}
	return x, nil
}

func ReadTranslationFile(path string) (*TS, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ts file: %v", path)
	}
	translationFile, err := readTranslationData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ts file: %v, %v", path, err)
	}
	translationFile.Path = path
	return translationFile, err
}

func ReadTranslationFiles(root, regxp string) (Translations, error) {
	m := Translations{}
	re := regexp.MustCompile(regxp)
	walkfunc := func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk through %s: %v", root, err)
		}
		if fileInfo.IsDir() || !re.MatchString(filepath.Base(path)) {
			return nil
		}
		translationFile, err := ReadTranslationFile(path)
		if err != nil {
			return err
		}
		m[translationFile.Lang] = translationFile
		return nil
	}
	err := filepath.Walk(root, walkfunc)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func ReadAllTranslationFiles(root string) (Translations, error) {
	return ReadTranslationFiles(root, `.*\.ts`)
}

func ReadTranslationFilesByPath(root, regxp string) (Translations, error) {
	m := Translations{}
	re := regexp.MustCompile(regxp)
	walkfunc := func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk through %s: %v", root, err)
		}
		if fileInfo.IsDir() || !re.MatchString(filepath.Base(path)) {
			return nil
		}
		translationFile, err := ReadTranslationFile(path)
		if err != nil {
			return err
		}
		m[translationFile.Path] = translationFile
		return nil
	}
	err := filepath.Walk(root, walkfunc)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (ts *Translations) Marshal(lang string) ([]byte, error) {
	if lang != "" {
		tsCopy := Translations{}
		tsCopy[lang] = (*ts)[lang]
		data, err := json.Marshal(tsCopy)
		if err != nil {
			return []byte{}, err
		}
		return data, nil
	}
	return json.Marshal(ts)
}

func (ts *Translations) MarshalIndent(lang string) ([]byte, error) {
	if lang != "" {
		tsCopy := Translations{}
		tsCopy[lang] = (*ts)[lang]
		data, err := json.MarshalIndent(tsCopy, "", "    ")
		if err != nil {
			return []byte{}, err
		}
		return data, nil
	}
	return json.MarshalIndent(ts, "", "    ")
}
