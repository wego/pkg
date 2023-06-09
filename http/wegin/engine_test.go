package wegin_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/http/wegin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testStruct struct {
	Reference string `json:"reference" binding:"required,alphanum_with_underscore_or_dash"`
}

func testHandler(c *gin.Context) {
	var b testStruct
	if err := c.ShouldBindJSON(&b); err == nil {
		c.JSON(http.StatusOK, gin.H{"reference": b.Reference})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

var (
	testEndpoint = "/test"
)

type EngineSuite struct {
	suite.Suite
	router *gin.Engine
}

func TestHandlers(t *testing.T) {
	suite.Run(t, new(EngineSuite))
}

// SetupTest runs before each Test
func (s *EngineSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = wegin.New()
	s.router.POST(testEndpoint, testHandler)
}

func (s *EngineSuite) Test_AlphaNumWithUnderscoreOrDash() {
	for _, testCase := range []struct {
		name           string
		referenceKey   string
		referenceValue string
		expectedStatus int
		expectedBodies []string
	}{
		{
			name:           "empty",
			referenceKey:   "reference",
			referenceValue: "",
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'required' tag",
			},
		},
		{
			name:           "all space",
			referenceKey:   "reference",
			referenceValue: "   ",
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'alphanum_with_underscore_or_dash' tag",
			},
		},
		{
			name:           "alphanum with space",
			referenceKey:   "reference",
			referenceValue: "abc def",
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'alphanum_with_underscore_or_dash' tag",
			},
		},
		{
			name:           "alphanum with underscore",
			referenceKey:   "reference",
			referenceValue: "abc123_def",
			expectedStatus: http.StatusOK,
			expectedBodies: []string{
				`"reference":"abc123_def"`,
			},
		},
		{
			name:           "alphanum with dash",
			referenceKey:   "reference",
			referenceValue: "abc123-def",
			expectedStatus: http.StatusOK,
			expectedBodies: []string{
				`"reference":"abc123-def"`,
			},
		},
		{
			name:           "alphanum with underscore and dash",
			referenceKey:   "reference",
			referenceValue: "abc123_def-ghi",
			expectedStatus: http.StatusOK,
			expectedBodies: []string{
				`"reference":"abc123_def-ghi"`,
			},
		},
	} {
		s.Run(testCase.name, func() {
			req, err := http.NewRequest(http.MethodPost, testEndpoint, strings.NewReader(fmt.Sprintf(`{"%s":"%s"}`, testCase.referenceKey, testCase.referenceValue)))
			s.NoError(err)

			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(testCase.expectedStatus, r.Code)
			for _, body := range testCase.expectedBodies {
				s.Contains(r.Body.String(), body)
			}
		})
	}
}
