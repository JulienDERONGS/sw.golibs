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
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ExtractedTranslation struct {
	Context string
	Sources []string
}

type ExtractedTranslations []*ExtractedTranslation

func (et ExtractedTranslations) Len() int {
	return len(et)
}

func (et ExtractedTranslations) Swap(i, j int) {
	et[i], et[j] = et[j], et[i]
}

func (et ExtractedTranslations) Less(i, j int) bool {
	return et[i].Context < et[j].Context
}

func applyRegexp(pattern, data string) ([]string, error) {
	regxp, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create regexp: %v", pattern)
	}
	matches := regxp.FindAllStringSubmatch(data, -1)
	result := []string{}
	for _, match := range matches {
		if len(match) >= 3 {
			result = append(result, match[2])
		}
	}
	return result, nil
}

func escapeString(str string) string {
	// Use &apos; and &aquot; instead of &#39; and &#34; to be consistent with qt.
	result := strings.ReplaceAll(html.EscapeString(str), "&#39;", "&apos;")
	result = strings.ReplaceAll(result, "&#34;", "&quot;")
	return result
}

func removeDuplicate(slice []string) []string {
	keys := make(map[string]bool)
	for _, entry := range slice {
		keys[escapeString(entry)] = true
	}
	list := []string{}
	for entry := range keys {
		list = append(list, entry)
	}
	sort.Strings(list)
	return list
}

func extractTranslation(data string, patterns []string) ([]string, error) {
	result := []string{}
	for _, pattern := range patterns {
		subResult, err := applyRegexp(pattern, data)
		if err != nil {
			return nil, err
		}
		result = append(result, subResult...)
	}
	return result, nil
}

func ExtractTranslations(root string, patterns []string,
	extensions, excludeDirs map[string]struct{}) (ExtractedTranslations, error) {

	result := ExtractedTranslations{}
	walkfunc := func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk through %s: %v", root, err)
		}
		_, ok := excludeDirs[filepath.Base(path)]
		if fileInfo.IsDir() && ok {
			return filepath.SkipDir
		}
		if fileInfo.IsDir() {
			return nil
		}
		_, ok = extensions[filepath.Ext(path)]
		if !ok {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", path)
		}
		matches, err := extractTranslation(string(data), patterns)
		if err != nil {
			return err
		}
		result = append(result, &ExtractedTranslation{
			Context: filepath.Base(path),
			Sources: removeDuplicate(matches),
		})
		return nil
	}
	err := filepath.Walk(root, walkfunc)
	if err != nil {
		return nil, err
	}
	sort.Sort(result)
	return result, nil
}

func BuildNewTranslationFiles(actualTranslations Translations,
	extractedTranslations ExtractedTranslations) (Translations, error) {

	result := Translations{}
	for lang, value := range actualTranslations {
		newTranslationFile := &TS{
			Lang:           value.Lang,
			Version:        value.Version,
			SourceLanguage: value.SourceLanguage,
			Path:           value.Path,
		}

		for _, extractedTranslation := range extractedTranslations {
			newContext := &Context{Name: extractedTranslation.Context}
			if len(extractedTranslation.Sources) == 0 {
				continue
			}

			for _, extractedSource := range extractedTranslation.Sources {
				newTranslation := &Translation{Type: "unfinished"}
				actualTranslation := value.FindTranslation(extractedTranslation.Context, extractedSource)
				if actualTranslation != nil { // If translation already exist under the same context, keep it
					newTranslation = actualTranslation
				}
				newContext.Messages = append(newContext.Messages,
					&Message{
						Source: &Source{
							Text: extractedSource,
						},
						Translation: newTranslation,
					})
			}
			sort.Sort(newContext.Messages) // not needed, but cleaner
			newTranslationFile.Contexts = append(newTranslationFile.Contexts, newContext)
		}
		result[lang] = newTranslationFile
	}
	return result, nil
}
