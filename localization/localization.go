package localization

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	fileExtension = ".json"
)

var (
	localizers sync.Map
)

// RegisterLocale Locales
func RegisterLocale(locale string, localeDir string) {
	bundle := i18n.NewBundle(language.MustParse(DefaultLocale))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	locale = strings.ToLower(locale)
	path := filepath.Join(localeDir, locale+fileExtension)
	bundle.MustLoadMessageFile(path)

	localizers.Store(locale, i18n.NewLocalizer(bundle, locale))
}

// RegisterLocales Locales
func RegisterLocales(locales []string, localeDir string) {
	for _, locale := range locales {
		RegisterLocale(locale, localeDir)
	}
}

func Localize(lang string, msgID string, data interface{}) (string, error) {
	lang = strings.ToLower(lang)
	localizer, ok := localizers.Load(lang)
	if !ok {
		// should always be OK, ignore the return value
		localizer, _ = localizers.Load(DefaultLocale)
	}

	return localizer.(*i18n.Localizer).Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
	})
}

func LocalizeMulti(lang string, msgIDs []string, data interface{}) (localizedMessages map[string]string, err error) {
	localizedMessages = make(map[string]string, len(msgIDs))
	for _, id := range msgIDs {
		msg, e := Localize(lang, id, data)
		if e != nil {
			return nil, e
		}
		localizedMessages[id] = msg
	}
	return
}
