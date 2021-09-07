//go:generate mockery --name=Service

package localization

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	fileExtension = ".json"
)

// Service ...
type Service interface {
	// Localize localizes a message & bind he data into message template
	Localize(lang string, msgID string, data interface{}) (string, error)
	// LocalizeMulti localizes multiple messages & return a map from message ID to the localized message.
	// It returns an error when any of the message ID cannot be localized.
	LocalizeMulti(lang string, msgIDs []string, data interface{}) (map[string]string, error)
}

type service struct {
	localizers map[string]*i18n.Localizer
}

// NewService initializes & returns a new localization.Service
func NewService(locales []string, localeDir string) Service {
	localizers := make(map[string]*i18n.Localizer, len(locales))
	bundle := i18n.NewBundle(language.MustParse(DefaultLocale))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, lang := range locales {
		lang = strings.ToLower(lang)
		path := filepath.Join(localeDir, lang+fileExtension)
		bundle.MustLoadMessageFile(path)

		localizers[lang] = i18n.NewLocalizer(bundle, lang)
	}

	return &service{localizers}
}

func (s *service) Localize(lang string, msgID string, data interface{}) (string, error) {
	lang = strings.ToLower(lang)
	localizer, ok := s.localizers[lang]
	if !ok {
		localizer = s.localizers[DefaultLocale]
	}

	return localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
	})
}

func (s *service) LocalizeMulti(lang string, msgIDs []string, data interface{}) (localizedMsgs map[string]string, err error) {
	localizedMsgs = make(map[string]string, len(msgIDs))
	for _, id := range msgIDs {
		msg, e := s.Localize(lang, id, data)
		if e != nil {
			return nil, e
		}
		localizedMsgs[id] = msg
	}
	return
}
