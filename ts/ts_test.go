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

func TestSortMessages(t *testing.T) {
	messages := Messages{
		&Message{Source: &Source{Text: "Z"}},
		&Message{Source: &Source{Text: "B"}},
		&Message{Source: &Source{Text: "A"}},
	}

	sort.Sort(messages)
	assert.Equal(t, messages[0].Source.Text, "A")
	assert.Equal(t, messages[1].Source.Text, "B")
	assert.Equal(t, messages[2].Source.Text, "Z")
}

var (
	simpleTs = &TS{
		Lang:           "en",
		SourceLanguage: "en",
		Version:        "1.0",
		Contexts: []*Context{
			&Context{
				Name: "context A",
				Messages: []*Message{
					&Message{
						Source:      &Source{Text: "&ltA&gt"},
						Translation: &Translation{Text: "tA"},
					},
					&Message{
						Source: &Source{Text: "B"},
						Translation: &Translation{
							Text: "tB",
							Type: "unfinished",
						},
					},
				},
			},
			&Context{
				Name: "context B",
				Messages: []*Message{
					&Message{
						Source:      &Source{Text: "C"},
						Translation: &Translation{Text: "tC"},
					},
				},
			},
		},
	}

	simpleTs2 = &TS{Lang: "fr", Version: "2.0", SourceLanguage: "en", Contexts: []*Context{
		&Context{Name: "Context1", Messages: []*Message{
			{Source: &Source{Text: "A"}, Translation: &Translation{Text: "B"}},
		}},
		&Context{Name: "Context2", Messages: []*Message{
			{Source: &Source{Text: "C"}, Translation: &Translation{Text: "D"}},
			{Source: &Source{Text: "E"}, Translation: &Translation{Text: "F", Type: "unfinished"}},
		}},
	}}

	simpleTr = Translations{
		"en": simpleTs,
		"fr": &TS{
			Lang: "fr",
			Contexts: []*Context{
				&Context{
					Name: "context A",
					Messages: []*Message{
						&Message{
							Source:      &Source{Text: "fr"},
							Translation: &Translation{Text: "tfr"},
						},
					},
				},
			},
		},
	}

	translationFileSpecialAttributes = TS{Lang: "fr", Version: "2.0", SourceLanguage: "en", Contexts: []*Context{
		&Context{Name: "Context", Messages: []*Message{
			{Source: &Source{Text: "km²"}, TranslatorComment: &TranslatorComment{"translator comment"}, Translation: &Translation{Text: "km²"}, Id: "1", OldSource: &OldSource{"old source"}, Utf8: true},
		}},
	}}
)

func TestTsFindTranslation(t *testing.T) {
	unknown := simpleTs.FindTranslation("unknown", "unknown")
	assert.Nil(t, unknown)
	unknown = simpleTs.FindTranslation("context A", "C")
	assert.Nil(t, unknown)

	tr := simpleTs.FindTranslation("context A", "B")
	assert.NotNil(t, tr)
	assert.Equal(t, tr.Text, "tB")
	tr = simpleTs.FindTranslation("context B", "C")
	assert.NotNil(t, tr)
	assert.Equal(t, tr.Text, "tC")

	tr = simpleTs.FindTranslation("context A", "<A>")
	assert.NotNil(t, tr)
	assert.Equal(t, tr.Text, "tA")
}

func TestTsFormat(t *testing.T) {
	value, err := simpleTs.Format()
	assert.Nil(t, err)
	assert.Equal(t, value,
		`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="1.0" language="en" sourcelanguage="en">
<context>
    <name>context A</name>
    <message>
        <source>&ltA&gt</source>
        <translation>tA</translation>
    </message>
    <message>
        <source>B</source>
        <translation type="unfinished">tB</translation>
    </message>
</context>
<context>
    <name>context B</name>
    <message>
        <source>C</source>
        <translation>tC</translation>
    </message>
</context>
</TS>
`)
}

func TestApplyTranslations(t *testing.T) {
	langs := map[string]struct{}{}
	messages := map[string]struct{}{}
	simpleTr.Apply(func(lang, context string, msg *Message) {
		langs[lang] = struct{}{}
		messages[msg.Source.Text] = struct{}{}
	})

	assert.Len(t, langs, 2)
	assert.Equal(t, langs, map[string]struct{}{
		"fr": struct{}{},
		"en": struct{}{},
	})
	assert.Len(t, messages, 4)
	assert.Equal(t, messages, map[string]struct{}{
		"&ltA&gt": struct{}{},
		"B":       struct{}{},
		"C":       struct{}{},
		"fr":      struct{}{},
	})
}

const (
	data = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="2.0" language="fr" sourcelanguage="en">
<context>
    <name>Context1</name>
    <message>
        <source>A</source>
        <translation>B</translation>
    </message>
</context>
<context>
    <name>Context2</name>
    <message>
        <source>C</source>
        <translation>D</translation>
    </message>
    <message>
        <source>E</source>
        <translation type="unfinished">F</translation>
    </message>
</context>
</TS>
`
	bytesData = `{
    "test": {
        "Lang": "fr",
        "SourceLanguage": "en",
        "Version": "2.0",
        "Contexts": [
            {
                "Name": "Context1",
                "Messages": [
                    {
                        "Numerus": "",
                        "Source": {
                            "Text": "A"
                        },
                        "TranslatorComment": null,
                        "Translation": {
                            "Type": "",
                            "Text": "B"
                        },
                        "Id": "",
                        "OldSource": null,
                        "Utf8": false
                    }
                ]
            },
            {
                "Name": "Context2",
                "Messages": [
                    {
                        "Numerus": "",
                        "Source": {
                            "Text": "C"
                        },
                        "TranslatorComment": null,
                        "Translation": {
                            "Type": "",
                            "Text": "D"
                        },
                        "Id": "",
                        "OldSource": null,
                        "Utf8": false
                    },
                    {
                        "Numerus": "",
                        "Source": {
                            "Text": "E"
                        },
                        "TranslatorComment": null,
                        "Translation": {
                            "Type": "unfinished",
                            "Text": "F"
                        },
                        "Id": "",
                        "OldSource": null,
                        "Utf8": false
                    }
                ]
            }
        ]
    }
}`

	specialAttributes = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE TS>
<TS version="2.0" language="fr" sourcelanguage="en">
<context>
	<name>Context</name>
	<message utf8="true" id="1">
		<source>km²</source>
		<oldsource>old source</oldsource>
		<translatorcomment>translator comment</translatorcomment>
		<translation>km²</translation>
	</message>
</context>
</TS>
`
)

func TestParsingTranslation(t *testing.T) {
	generatedResult, err := readTranslationData([]byte(data))
	assert.NoError(t, err)
	assert.Equal(t, simpleTs2, generatedResult)
}

func TestTranslationWriting(t *testing.T) {
	value, err := simpleTs2.Format()
	assert.NoError(t, err)
	assert.Equal(t, value, data)
}

func TestMarshalTs(t *testing.T) {
	var datas Translations = make(map[string]*TS)
	datas["test"] = simpleTs2
	result, err := datas.MarshalIndent("")
	assert.NoError(t, err)
	assert.Equal(t, bytesData, string(result))
}

func TestSpecialAttributes(t *testing.T) {
	generatedResult, err := readTranslationData([]byte(specialAttributes))
	assert.NoError(t, err)
	assert.Equal(t, translationFileSpecialAttributes, *generatedResult)
}
