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
	Reference        string  `json:"reference" binding:"required,alphanum_with_underscore_or_dash"`
	ReferencePointer *string `json:"reference_pointer" binding:"required,alphanum_with_underscore_or_dash"`
}

func testHandler(c *gin.Context) {
	var b testStruct
	if err := c.ShouldBindJSON(&b); err == nil {
		c.JSON(http.StatusOK, b)
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
		name            string
		refKey          string
		refValue        string
		refPointerKey   string
		refPointerValue string
		expectedStatus  int
		expectedBodies  []string
	}{
		{
			name:           "empty",
			refKey:         "reference",
			refValue:       "",
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'required' tag",
				"Key: 'testStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'required' tag",
			},
		},
		{
			name:            "empty value",
			refPointerKey:   "reference_pointer",
			refPointerValue: "",
			expectedStatus:  http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'required' tag",
				"Key: 'testStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'alphanum_with_underscore_or_dash' tag",
			},
		},
		{
			name:            "all space",
			refKey:          "reference",
			refValue:        "   ",
			refPointerKey:   "reference_pointer",
			refPointerValue: "   ",
			expectedStatus:  http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'alphanum_with_underscore_or_dash' tag",
				"Key: 'testStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'alphanum_with_underscore_or_dash' tag",
			},
		},
		{
			name:            "alphanum with space",
			refKey:          "reference",
			refValue:        "abc def",
			refPointerKey:   "reference_pointer",
			refPointerValue: "abc def",
			expectedStatus:  http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'testStruct.Reference' Error:Field validation for 'Reference' failed on the 'alphanum_with_underscore_or_dash' tag",
				"Key: 'testStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'alphanum_with_underscore_or_dash' tag",
			},
		},
		{
			name:            "alphanum with underscore",
			refKey:          "reference",
			refValue:        "abc123_def",
			refPointerKey:   "reference_pointer",
			refPointerValue: "abc123_def",
			expectedStatus:  http.StatusOK,
			expectedBodies: []string{
				`"reference":"abc123_def"`,
				`"reference_pointer":"abc123_def"`,
			},
		},
		{
			name:            "alphanum with dash",
			refKey:          "reference",
			refValue:        "abc123-def",
			refPointerKey:   "reference_pointer",
			refPointerValue: "abc123-def",
			expectedStatus:  http.StatusOK,
			expectedBodies: []string{
				`"reference":"abc123-def"`,
				`"reference_pointer":"abc123-def"`,
			},
		},
		{
			name:            "alphanum with underscore and dash",
			refKey:          "reference",
			refValue:        "abc123_def-ghi",
			refPointerKey:   "reference_pointer",
			refPointerValue: "abc123_def-ghi",
			expectedStatus:  http.StatusOK,
			expectedBodies: []string{
				`"reference":"abc123_def-ghi"`,
				`"reference_pointer":"abc123_def-ghi"`,
			},
		},
	} {
		s.Run(testCase.name, func() {
			req, err := http.NewRequest(http.MethodPost, testEndpoint,
				strings.NewReader(fmt.Sprintf(`{"%s":"%s", "%s": "%s"}`,
					testCase.refKey, testCase.refValue, testCase.refPointerKey, testCase.refPointerValue)))
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
