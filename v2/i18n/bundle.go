package i18n

import (
	"fmt"
	"os"

	"github.com/abdullahskartal/go-i18n/v2/internal/plural"

	"golang.org/x/text/language"
)

// UnmarshalFunc unmarshals data into v.
type UnmarshalFunc func(data []byte, v interface{}) error

// Bundle stores a set of messages and pluralization rules.
// Most applications only need a single bundle
// that is initialized early in the application's lifecycle.
// It is not goroutine safe to modify the bundle while Localizers
// are reading from it.
type Bundle struct {
	defaultLanguage  language.Tag
	unmarshalFuncs   map[string]UnmarshalFunc
	messageTemplates map[string]map[string]*MessageTemplate
	pluralRules      plural.Rules
	countryTagPair   map[string][]language.Tag
	matcher          language.Matcher
}

// artTag is the language tag used for artificial languages
// https://en.wikipedia.org/wiki/Codes_for_constructed_languages
var artTag = language.MustParse("art")

// NewBundle returns a bundle with a default language and a default set of plural rules.
func NewBundle(countryCode string, defaultLanguage language.Tag) *Bundle {
	b := &Bundle{
		defaultLanguage: defaultLanguage,
		pluralRules:     plural.DefaultRules(),
	}
	b.pluralRules[artTag] = b.pluralRules.Rule(language.English)
	b.addTag(countryCode, defaultLanguage)
	return b
}

// RegisterUnmarshalFunc registers an UnmarshalFunc for format.
func (b *Bundle) RegisterUnmarshalFunc(format string, unmarshalFunc UnmarshalFunc) {
	if b.unmarshalFuncs == nil {
		b.unmarshalFuncs = make(map[string]UnmarshalFunc)
	}
	b.unmarshalFuncs[format] = unmarshalFunc
}

// LoadMessageFile loads the bytes from path
// and then calls ParseMessageFileBytes.
func (b *Bundle) LoadMessageFile(path, countryCode string) (*MessageFile, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b.ParseMessageFileBytes(buf, path, countryCode)
}

// MustLoadMessageFile is similar to LoadMessageFile
// except it panics if an error happens.
func (b *Bundle) MustLoadMessageFile(countryCode, path string) {
	if _, err := b.LoadMessageFile(path, countryCode); err != nil {
		panic(err)
	}
}

// ParseMessageFileBytes parses the bytes in buf to add translations to the bundle.
//
// The format of the file is everything after the last ".".
//
// The language tag of the file is everything after the second to last "." or after the last path separator, but before the format.
func (b *Bundle) ParseMessageFileBytes(buf []byte, path, countryCode string) (*MessageFile, error) {
	messageFile, err := ParseMessageFileBytes(buf, path, b.unmarshalFuncs)
	if err != nil {
		return nil, err
	}
	if err := b.AddMessages(countryCode, messageFile.Tag, messageFile.Messages...); err != nil {
		return nil, err
	}
	return messageFile, nil
}

// MustParseMessageFileBytes is similar to ParseMessageFileBytes
// except it panics if an error happens.
func (b *Bundle) MustParseMessageFileBytes(buf []byte, path, countryCode string) {
	if _, err := b.ParseMessageFileBytes(buf, path, countryCode); err != nil {
		panic(err)
	}
}

// AddMessages adds messages for a language.
// It is useful if your messages are in a format not supported by ParseMessageFileBytes.
func (b *Bundle) AddMessages(countryCode string, tag language.Tag, messages ...*Message) error {
	pluralRule := b.pluralRules.Rule(tag)
	if pluralRule == nil {
		return fmt.Errorf("no plural rule registered for %s", tag)
	}
	key := b.messageTemplateKey(countryCode, tag)
	if b.messageTemplates == nil {
		b.messageTemplates = map[string]map[string]*MessageTemplate{}
	}
	if b.messageTemplates[key] == nil {
		b.messageTemplates[key] = map[string]*MessageTemplate{}
		b.addTag(countryCode, tag)
	}
	for _, m := range messages {
		b.messageTemplates[key][m.ID] = NewMessageTemplate(m)
	}
	return nil
}

// MustAddMessages is similar to AddMessages except it panics if an error happens.
func (b *Bundle) MustAddMessages(countryCode string, tag language.Tag, messages ...*Message) {
	if err := b.AddMessages(countryCode, tag, messages...); err != nil {
		panic(err)
	}
}

func (b *Bundle) addTag(countryCode string, tag language.Tag) {
	for cc, tags := range b.countryTagPair {
		for _, t := range tags {
			if cc == countryCode && t == tag {
				// Tag already exists
				return
			}
		}
	}

	if b.countryTagPair == nil {
		b.countryTagPair = make(map[string][]language.Tag)
	}
	b.countryTagPair[countryCode] = append(b.countryTagPair[countryCode], tag)
	b.matcher = language.NewMatcher(b.countryTagPair[countryCode])
}

// LanguageTags returns the list of language tags
// of all the translations loaded into the bundle
func (b *Bundle) LanguageTags(countryCode string) []language.Tag {
	return b.countryTagPair[countryCode]
}

func (b *Bundle) getMessageTemplate(tag language.Tag, id, countryCode string) *MessageTemplate {
	key := b.messageTemplateKey(countryCode, tag)
	templates := b.messageTemplates[key]
	if templates == nil {
		return nil
	}
	return templates[id]
}

func (b *Bundle) messageTemplateKey(countryCode string, tag language.Tag) string {
	return countryCode + "-" + tag.String()
}
