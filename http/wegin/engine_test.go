package wegin_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/http/wegin"
)

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

type alphaNumTestStruct struct {
	Reference        string  `json:"reference" binding:"required,alphanum_with_underscore_or_dash"`
	ReferencePointer *string `json:"reference_pointer" binding:"required,alphanum_with_underscore_or_dash"`
	OneOf            *string `json:"one_of_pointer" binding:"omitempty,one_of_or_blank=foo bar"`
}

// New struct for testing one_of_or_blank with different types
type oneOfOrBlankTestStruct struct {
	// Only support string pointer type
	StringPtrField1 *string `json:"string_ptr_field1" binding:"omitempty,one_of_or_blank=red blue green"`
	StringPtrField2 *string `json:"string_ptr_field2" binding:"omitempty,one_of_or_blank=apple orange banana"`
	StringPtrField3 *string `json:"string_ptr_field3" binding:"omitempty,one_of_or_blank=dog cat fish"`
}

func alphaNumHandler(c *gin.Context) {
	var b alphaNumTestStruct
	if err := c.ShouldBindJSON(&b); err == nil {
		c.JSON(http.StatusOK, b)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func oneOfOrBlankTestHandler(c *gin.Context) {
	var b oneOfOrBlankTestStruct
	if err := c.ShouldBindJSON(&b); err == nil {
		c.JSON(http.StatusOK, b)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

var (
	alphaNumEndpoint     = "/test-alphanum-with-dash"
	oneOfOrBlankEndpoint = "/test-one-of-or-blank"
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
	s.router.POST(alphaNumEndpoint, alphaNumHandler)
	s.router.POST(oneOfOrBlankEndpoint, oneOfOrBlankTestHandler)
}

func (s *EngineSuite) Test_AlphaNumWithUnderscoreOrDash() {
	testCases := []struct {
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
				"Key: 'alphaNumTestStruct.Reference' Error:Field validation for 'Reference' failed on the 'required' tag",
				"Key: 'alphaNumTestStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'required' tag",
			},
		},
		{
			name:            "empty value",
			refPointerKey:   "reference_pointer",
			refPointerValue: "",
			expectedStatus:  http.StatusBadRequest,
			expectedBodies: []string{
				"Key: 'alphaNumTestStruct.Reference' Error:Field validation for 'Reference' failed on the 'required' tag",
				"Key: 'alphaNumTestStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'alphanum_with_underscore_or_dash' tag",
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
				"Key: 'alphaNumTestStruct.Reference' Error:Field validation for 'Reference' failed on the 'alphanum_with_underscore_or_dash' tag",
				"Key: 'alphaNumTestStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'alphanum_with_underscore_or_dash' tag",
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
				"Key: 'alphaNumTestStruct.Reference' Error:Field validation for 'Reference' failed on the 'alphanum_with_underscore_or_dash' tag",
				"Key: 'alphaNumTestStruct.ReferencePointer' Error:Field validation for 'ReferencePointer' failed on the 'alphanum_with_underscore_or_dash' tag",
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
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Build request body
			jsonBody := `{`
			if tc.refKey != "" {
				jsonBody += fmt.Sprintf(`"%s":"%s"`, tc.refKey, tc.refValue)
			}
			if tc.refPointerKey != "" {
				if tc.refKey != "" {
					jsonBody += ", "
				}
				jsonBody += fmt.Sprintf(`"%s":"%s"`, tc.refPointerKey, tc.refPointerValue)
			}
			jsonBody += "}"

			// Create and send request
			req, err := http.NewRequest(http.MethodPost, alphaNumEndpoint, strings.NewReader(jsonBody))
			s.NoError(err)

			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			// Verify response
			s.Equal(tc.expectedStatus, r.Code)
			for _, body := range tc.expectedBodies {
				s.Contains(r.Body.String(), body)
			}
		})
	}
}

func (s *EngineSuite) Test_OneOfOrBlank() {
	testCases := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedBodies []string
		// Expected struct for successful responses
		expectedStruct *oneOfOrBlankTestStruct
	}{
		// StringPtrField1 tests
		{
			name:           "string pointer 1 - nil",
			payload:        `{}`,
			expectedStatus: http.StatusOK,
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 1 - empty string",
			payload:        `{"string_ptr_field1":""}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field1":""`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: strPtr(""),
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 1 - valid value red",
			payload:        `{"string_ptr_field1":"red"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field1":"red"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: strPtr("red"),
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 1 - valid value blue",
			payload:        `{"string_ptr_field1":"blue"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field1":"blue"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: strPtr("blue"),
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 1 - valid value green",
			payload:        `{"string_ptr_field1":"green"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field1":"green"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: strPtr("green"),
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 1 - invalid value",
			payload:        `{"string_ptr_field1":"yellow"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{`Field validation for 'StringPtrField1' failed on the 'one_of_or_blank' tag`},
		},

		// StringPtrField2 tests
		{
			name:           "string pointer 2 - nil",
			payload:        `{}`,
			expectedStatus: http.StatusOK,
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 2 - empty string",
			payload:        `{"string_ptr_field2":""}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field2":""`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: strPtr(""),
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 2 - valid value apple",
			payload:        `{"string_ptr_field2":"apple"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field2":"apple"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: strPtr("apple"),
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 2 - valid value orange",
			payload:        `{"string_ptr_field2":"orange"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field2":"orange"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: strPtr("orange"),
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 2 - valid value banana",
			payload:        `{"string_ptr_field2":"banana"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field2":"banana"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: strPtr("banana"),
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 2 - invalid value",
			payload:        `{"string_ptr_field2":"grape"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{`Field validation for 'StringPtrField2' failed on the 'one_of_or_blank' tag`},
		},

		// StringPtrField3 tests
		{
			name:           "string pointer 3 - nil",
			payload:        `{}`,
			expectedStatus: http.StatusOK,
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: nil,
			},
		},
		{
			name:           "string pointer 3 - empty string",
			payload:        `{"string_ptr_field3":""}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field3":""`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: strPtr(""),
			},
		},
		{
			name:           "string pointer 3 - valid value dog",
			payload:        `{"string_ptr_field3":"dog"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field3":"dog"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: strPtr("dog"),
			},
		},
		{
			name:           "string pointer 3 - valid value cat",
			payload:        `{"string_ptr_field3":"cat"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field3":"cat"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: strPtr("cat"),
			},
		},
		{
			name:           "string pointer 3 - valid value fish",
			payload:        `{"string_ptr_field3":"fish"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{`"string_ptr_field3":"fish"`},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: nil,
				StringPtrField2: nil,
				StringPtrField3: strPtr("fish"),
			},
		},
		{
			name:           "string pointer 3 - invalid value",
			payload:        `{"string_ptr_field3":"bird"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{`Field validation for 'StringPtrField3' failed on the 'one_of_or_blank' tag`},
		},

		// Multiple fields test
		{
			name:           "multiple fields - all valid",
			payload:        `{"string_ptr_field1":"red", "string_ptr_field2":"apple", "string_ptr_field3":"dog"}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{
				`"string_ptr_field1":"red"`,
				`"string_ptr_field2":"apple"`,
				`"string_ptr_field3":"dog"`,
			},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: strPtr("red"),
				StringPtrField2: strPtr("apple"),
				StringPtrField3: strPtr("dog"),
			},
		},
		{
			name:           "multiple fields - one invalid",
			payload:        `{"string_ptr_field1":"red", "string_ptr_field2":"grape", "string_ptr_field3":"dog"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{`Field validation for 'StringPtrField2' failed on the 'one_of_or_blank' tag`},
		},
		{
			name:           "multiple fields - all empty",
			payload:        `{"string_ptr_field1":"", "string_ptr_field2":"", "string_ptr_field3":""}`,
			expectedStatus: http.StatusOK,
			expectedBodies: []string{
				`"string_ptr_field1":""`,
				`"string_ptr_field2":""`,
				`"string_ptr_field3":""`,
			},
			expectedStruct: &oneOfOrBlankTestStruct{
				StringPtrField1: strPtr(""),
				StringPtrField2: strPtr(""),
				StringPtrField3: strPtr(""),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create and send request
			req, err := http.NewRequest(http.MethodPost, oneOfOrBlankEndpoint, strings.NewReader(tc.payload))
			s.NoError(err)

			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			// Verify response
			s.Equal(tc.expectedStatus, r.Code)
			for _, body := range tc.expectedBodies {
				s.Contains(r.Body.String(), body)
			}

			// For successful requests, verify the struct values
			if tc.expectedStatus == http.StatusOK {
				var result oneOfOrBlankTestStruct
				err = json.Unmarshal(r.Body.Bytes(), &result)
				s.NoError(err)

				// Compare the entire struct
				if tc.expectedStruct != nil {
					// Check StringPtrField1
					if tc.expectedStruct.StringPtrField1 == nil {
						s.Nil(result.StringPtrField1, "StringPtrField1 should be nil")
					} else {
						s.NotNil(result.StringPtrField1, "StringPtrField1 should not be nil")
						s.Equal(*tc.expectedStruct.StringPtrField1, *result.StringPtrField1, "StringPtrField1 value mismatch")
					}

					// Check StringPtrField2
					if tc.expectedStruct.StringPtrField2 == nil {
						s.Nil(result.StringPtrField2, "StringPtrField2 should be nil")
					} else {
						s.NotNil(result.StringPtrField2, "StringPtrField2 should not be nil")
						s.Equal(*tc.expectedStruct.StringPtrField2, *result.StringPtrField2, "StringPtrField2 value mismatch")
					}

					// Check StringPtrField3
					if tc.expectedStruct.StringPtrField3 == nil {
						s.Nil(result.StringPtrField3, "StringPtrField3 should be nil")
					} else {
						s.NotNil(result.StringPtrField3, "StringPtrField3 should not be nil")
						s.Equal(*tc.expectedStruct.StringPtrField3, *result.StringPtrField3, "StringPtrField3 value mismatch")
					}
				}
			}
		})
	}
}
