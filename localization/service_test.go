package localization_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/localization"
)

type ServiceSuite struct {
	suite.Suite
	svc localization.Service
}

// SetupTest runs before each Test
func (s *ServiceSuite) SetupTest() {
	locales := []string{"en", "ar"}
	localeDir := "./testdata"
	s.svc = localization.NewService(locales, localeDir)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) Test_Localize_NotOk() {
	// localize non-exist key
	msg, err := s.svc.Localize("en", "i.am.supperman", nil)
	s.Error(err)
	s.Contains(err.Error(), "not found in language")
	s.Zero(msg)
}

func (s *ServiceSuite) Test_Localize_Ok() {
	// localize supported language
	msg, err := s.svc.Localize("ar", "mailer.terms_conditions", nil)
	s.NoError(err)
	s.Equal("الشروط والأحكام", msg)

	// localize non-supported language -> use default language
	msg, err = s.svc.Localize("vn", "mailer.refund_initiated.salutation", map[string]string{"CustomerFullName": "Aaron Ramsey"})
	s.NoError(err)
	s.Equal("Hi Aaron Ramsey", msg)
}

func (s *ServiceSuite) Test_LocalizeMulti_NotOk() {
	key1, key2 := "mailer.support_phone_worldwide", "mailer.footer_copyright"
	msgs, err := s.svc.LocalizeMulti("vn", []string{key1, key2}, map[string]string{"CurrentYear": "2021"})
	s.NoError(err)
	s.Equal("Worldwide ", msgs[key1])
	s.Equal("©2021 Wego Pte Ltd. All rights reserved.", msgs[key2])
}

func (s *ServiceSuite) Test_LocalizeMulti_Ok() {
	key1, key2 := "mailer.footer_privacy", "mailer.footer_copyright"
	msgs, err := s.svc.LocalizeMulti("vn", []string{key1, key2, "i.am.handsome"}, map[string]string{"CurrentYear": "2021"})
	s.Error(err)
	s.Contains(err.Error(), "not found in language")
	s.Nil(msgs)
}
