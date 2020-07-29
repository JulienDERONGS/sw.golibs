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
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestSortTranslations(t *testing.T) {
	translations := ExtractedTranslations{
		&ExtractedTranslation{Context: "Z"},
		&ExtractedTranslation{Context: "B"},
		&ExtractedTranslation{Context: "A"},
	}

	sort.Sort(translations)
	assert.Equal(t, translations[0].Context, "A")
	assert.Equal(t, translations[1].Context, "B")
	assert.Equal(t, translations[2].Context, "Z")
}

func TestExtractingRegexpMatches(t *testing.T) {
	expectedResult := []string{"test_begin", "test_middle", "test_end"}

	dataDoubleQuote := `
i18n "test_begin" bla bla bla bla bla bla bla bla bla bla
bla bla bla i18n "test_middle" bla bla bla bla bla bla bla
bla bla bla bla bla bla bla bla bla bla i18n "test_end"
`

	dataSimpleQuote := `
i18n 'test_begin' bla bla bla bla bla bla bla bla bla bla
bla bla bla i18n 'test_middle' bla bla bla bla bla bla bla
bla bla bla bla bla bla bla bla bla bla i18n 'test_end'
`

	matches, err := applyRegexp(`(i18n)\s+"([^"]+)"`, dataDoubleQuote)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, matches)

	matches, err = applyRegexp(`(i18n)\s+'([^']+)'`, dataSimpleQuote)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, matches)

	matches, err = extractTranslation(dataDoubleQuote+"\n"+dataSimpleQuote,
		[]string{`(i18n)\s+"([^"]+)"`, `(i18n)\s+'([^']+)'`})
	assert.NoError(t, err)
	assert.Equal(t, append(expectedResult, expectedResult...), matches)
}

func TestRemoveDuplicate(t *testing.T) {
	assert.Equal(t,
		removeDuplicate([]string{"a", "a", "b", "c"}),
		[]string{"a", "b", "c"})

	assert.Equal(t,
		removeDuplicate([]string{"<", ">", "c", ">"}),
		[]string{"&gt;", "&lt;", "c"})
}

func TestBuildingNewTranslationFile(t *testing.T) {
	translations := Translations{
		"en": &TS{Lang: "en", Version: "2.0", SourceLanguage: "en", Contexts: []*Context{
			&Context{Name: "Context1", Messages: []*Message{
				{Source: &Source{Text: "A"}, Translation: &Translation{Text: "B"}},
				{Source: &Source{Text: "B"}, Translation: &Translation{Text: "B"}},
			}},
			&Context{Name: "Context2", Messages: []*Message{
				{Source: &Source{Text: "D"}, Translation: &Translation{Text: "D"}},
				{Source: &Source{Text: "E"}, Translation: &Translation{Text: "E"}},
				{Source: &Source{Text: "F"}, Translation: &Translation{Text: "F"}},
				{Source: &Source{Text: "G"}, Translation: &Translation{Text: "G"}},
			}},
		}},
	}

	extractedTranslations := ExtractedTranslations{
		&ExtractedTranslation{Context: "Context1", Sources: []string{"A", "B", "C"}},
		&ExtractedTranslation{Context: "Context2", Sources: []string{"D", "E", "F"}},
	}

	expectedTranslations := Translations{
		"en": &TS{Lang: "en", Version: "2.0", SourceLanguage: "en", Contexts: []*Context{
			&Context{Name: "Context1", Messages: []*Message{
				{Source: &Source{Text: "A"}, Translation: &Translation{Text: "B"}},
				{Source: &Source{Text: "B"}, Translation: &Translation{Text: "B"}},
				{Source: &Source{Text: "C"}, Translation: &Translation{Type: "unfinished"}},
			}},
			&Context{Name: "Context2", Messages: []*Message{
				{Source: &Source{Text: "D"}, Translation: &Translation{Text: "D"}},
				{Source: &Source{Text: "E"}, Translation: &Translation{Text: "E"}},
				{Source: &Source{Text: "F"}, Translation: &Translation{Text: "F"}},
			}},
		}},
	}

	newTranslations, err := BuildNewTranslationFiles(translations, extractedTranslations)
	assert.NoError(t, err)
	assert.Equal(t, expectedTranslations, newTranslations)
}
