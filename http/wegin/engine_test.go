package wegin_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/http/wegin"
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

func valueIfHandler(c *gin.Context) {
	var p parentStruct
	if err := c.ShouldBindJSON(&p); err == nil {
		c.JSON(http.StatusOK, p)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

var (
	testEndpoint        = "/test"
	testValueIfEndpoint = "/test_value_if"
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
	s.router.POST(testValueIfEndpoint, valueIfHandler)
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

type parentStruct struct {
	Type  string       `json:"type" binding:"required,oneof=AA BB"`
	Child *childStruct `json:"child" binding:"required,dive"`
}

type childStruct struct {
	Type   string             `json:"type" binding:"required,oneof=aa bb"`
	Child  *grandChildStruct  `json:"child"`
	Child2 *grandChildStruct2 `json:"child2"`
}

type grandChildStruct struct {
	Name string `json:"name" binding:"oneof=AAaa11 BBbb22,value_if=root.Type AA == AAaa11"`
	Fee  uint32 `json:"fee" binding:"value_if=root.Type AA == 10"`
}

type grandChildStruct2 struct {
	Name string   `json:"name" binding:"oneof=AAaa11 BBbb22,value_if=root.Child.Type aa == BBbb22"`
	Rate *float32 `json:"rate" binding:"omitempty,value_if=root.Type AA == 0.2"`
}

func (s *EngineSuite) Test_ValueIf() {
	for _, testCase := range []struct {
		name           string
		body           string
		expectedStatus int
		expectedBodies []string
	}{
		{
			name:           "pass",
			body:           `{"type":"AA","child":{"type":"aa","child":{"name":"AAaa11","fee":10},"child2":{"name":"BBbb22","rate":0.2}}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "any value for fee if root.type is not AA should pass",
			body:           `{"type":"BB","child":{"type":"aa","child":{"name":"AAaa11","fee":20},"child2":{"name":"BBbb22"}}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "root.child.child.name must be AAaa11 if root.type is AA",
			body:           `{"type":"AA","child":{"type":"aa","child":{"name":"BBbb22","fee":10},"child2":{"name":"BBbb22"}}}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				`Key: 'parentStruct.Child.Child.Name' Error:Field validation for 'Name' failed on the 'value_if' tag`,
			},
		},
		{
			name:           "root.child.child2.name must be BBbb22 if root.child.type is aa",
			body:           `{"type":"AA","child":{"type":"aa","child":{"name":"AAaa11","fee":10},"child2":{"name":"AAaa11"}}}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				`Key: 'parentStruct.Child.Child2.Name' Error:Field validation for 'Name' failed on the 'value_if' tag`,
			},
		},
		{
			name:           "root.child.child.fee must be 10 if root.type is AA",
			body:           `{"type":"AA","child":{"type":"aa","child":{"name":"AAaa11","fee":20},"child2":{"name":"BBbb22"}}}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				`Key: 'parentStruct.Child.Child.Fee' Error:Field validation for 'Fee' failed on the 'value_if' tag`,
			},
		},
		{
			name:           "root.child.child.name can be any value if root.Type is not AA",
			body:           `{"type":"BB","child":{"type":"bb","child":{"name":"BBbb22","fee":10},"child2":{"name":"BBbb22"}}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "root.child.child2.name can be any value if root.child.type is not aa",
			body:           `{"type":"AA","child":{"type":"bb","child":{"name":"AAaa11","fee":10},"child2":{"name":"AAaa11"}}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "root.child.child2.rate must be 0.2 if root.type is AA",
			body:           `{"type":"AA","child":{"type":"aa","child":{"name":"AAaa11","fee":10},"child2":{"name":"BBbb22","rate":0.3}}}`,
			expectedStatus: http.StatusBadRequest,
			expectedBodies: []string{
				`Key: 'parentStruct.Child.Child2.Rate' Error:Field validation for 'Rate' failed on the 'value_if' tag`,
			},
		},
		{
			name:           "root.child.child2.rate can be any value if root.type is not AA",
			body:           `{"type":"BB","child":{"type":"bb","child":{"name":"BBbb22","fee":10},"child2":{"name":"BBbb22","rate":0.3}}}`,
			expectedStatus: http.StatusOK,
		},
	} {
		s.Run(testCase.name, func() {
			req, err := http.NewRequest(http.MethodPost, testValueIfEndpoint, strings.NewReader(testCase.body))
			s.NoError(err)

			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(testCase.expectedStatus, r.Code)

			if testCase.expectedBodies != nil {
				for _, body := range testCase.expectedBodies {
					s.Contains(r.Body.String(), body)
				}
			}
		})
	}
}
