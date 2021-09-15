package localization_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/localization"
)

type LocalizationSuite struct {
	suite.Suite
}

// SetupTest runs before each Test
func (s *LocalizationSuite) SetupTest() {
	locales := []string{"en", "ar"}
	localeDir := "./testdata"
	localization.RegisterLocales(locales, localeDir)
}

func TestLocalizationSuite(t *testing.T) {
	suite.Run(t, new(LocalizationSuite))
}

func (s *LocalizationSuite) Test_Localize_NotOk() {
	// localize non-exist key
	msg, err := localization.Localize("en", "i.am.superman", nil)
	s.Error(err)
	s.Contains(err.Error(), "not found in language")
	s.Zero(msg)
}

func (s *LocalizationSuite) Test_Localize_Ok() {
	// localize supported language
	msg, err := localization.Localize("ar", "mailer.terms_conditions", nil)
	s.NoError(err)
	s.Equal("الشروط والأحكام", msg)

	// localize non-supported language -> use default language
	msg, err = localization.Localize("vn", "mailer.refund_initiated.salutation", map[string]string{"CustomerFullName": "Aaron Ramsey"})
	s.NoError(err)
	s.Equal("Hi Aaron Ramsey", msg)
}

func (s *LocalizationSuite) Test_LocalizeMulti_NotOk() {
	key1, key2 := "mailer.support_phone_worldwide", "mailer.footer_copyright"
	msgs, err := localization.LocalizeMulti("vn", []string{key1, key2}, map[string]string{"CurrentYear": "2021"})
	s.NoError(err)
	s.Equal("Worldwide ", msgs[key1])
	s.Equal("©2021 Wego Pte Ltd. All rights reserved.", msgs[key2])
}

func (s *LocalizationSuite) Test_LocalizeMulti_Ok() {
	key1, key2 := "mailer.footer_privacy", "mailer.footer_copyright"
	msgs, err := localization.LocalizeMulti("vn", []string{key1, key2, "i.am.handsome"}, map[string]string{"CurrentYear": "2021"})
	s.Error(err)
	s.Contains(err.Error(), "not found in language")
	s.Nil(msgs)
}
