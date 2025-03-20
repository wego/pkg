package binding_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/audit"
	"github.com/wego/pkg/errors"
	"github.com/wego/pkg/http/binding"
)

// Helper function to create string pointers
func pointer[T any](v T) *T {
	return &v
}

var (
	testJSONEndpoint          = "/test/json"
	testMultipleBindEndpoint  = "/test/multiple"
	testQueryEndpoint         = "/test/query"
	testChangeRequestEndpoint = "/test/cr"
	testIDEndpoint            = "/test"
	testUriEndpoint           = "/test/:id/child/:child_id"
	ctxKey                    = "testRequest"
	anotherCtxKey             = "anotherTestRequest"
)

type testStruct struct {
	Number []uint   `form:"Num,omitempty" json:"number" binding:"required,dive,number,min=1"`
	String []string `form:"Str,omitempty" json:"string" binding:"required,dive,printascii"`
}

type anotherTestStruct struct {
	Number []uint   `form:"Num,omitempty" json:"number" binding:"required,dive,number,min=1"`
	String []string `form:"Str,omitempty" json:"string" binding:"required,dive,printascii"`
	Dummy  string   `form:"Dummy,omitempty" json:"dummy" binding:"required"`
}

type testChangeStruct struct {
	Number []uint   `form:"Num,omitempty" json:"number" binding:"required,dive,number,min=1"`
	String []string `form:"Str,omitempty" json:"string" binding:"required,dive,printascii"`
	audit.ChangeRequest
}

type testUriStruct struct {
	ID      string `uri:"id" json:"id"`
	ChildID string `uri:"child_id" json:"child_id"`
	Name    string `json:"name"`
}

func shouldBindJSONHandler(c *gin.Context) {
	var t testStruct
	if err := binding.ShouldBindJSON(c, ctxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func bindJSONHandler(c *gin.Context) {
	var t testStruct
	if err := binding.BindJSON(c, ctxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func aBindJSONHandler(c *gin.Context) {
	var t testStruct
	if err := binding.BindJSON(c, ctxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	return
}

func anotherBindJSONHandler(c *gin.Context) {
	var t anotherTestStruct
	if err := binding.BindJSON(c, anotherCtxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func bindQueryHandler(c *gin.Context) {
	var t testStruct
	if err := binding.BindQuery(c, ctxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func bindChangeRequestHandler(c *gin.Context) {
	var t testChangeStruct
	if err := binding.BindChangeRequest(c, ctxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func bindIDHandler(c *gin.Context) {
	id, err := binding.BindID(c)
	if err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, id)
}

func bindUriHandler(c *gin.Context) {
	var t testUriStruct
	ctxKeyUri := "keyUri"
	ctxKeyBody := "keyBody"

	if err := binding.BindURI(c, ctxKeyUri, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	if err := binding.BindJSON(c, ctxKeyBody, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
		return
	}
	c.JSON(http.StatusOK, t)
}

type BindingSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupTest runs before each Test
func (s *BindingSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
}

func TestHandlers(t *testing.T) {
	suite.Run(t, new(BindingSuite))
}

func (s *BindingSuite) Test_ShouldBindJSON() {
	testCases := []struct {
		name           string
		method         string
		requestBody    string
		setupContext   bool
		expectedStatus int
		expectedBody   string
		verifyEmpty    bool
	}{
		{
			name:           "FromContext",
			method:         http.MethodPatch,
			requestBody:    `{"number":[1],"string":["A"]}`,
			setupContext:   true,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"]}`,
		},
		{
			name:           "NoBody",
			method:         http.MethodPatch,
			requestBody:    "",
			setupContext:   false,
			expectedStatus: http.StatusOK,
			verifyEmpty:    true,
		},
		{
			name:           "FromBody",
			method:         http.MethodPatch,
			requestBody:    `{"number":[1,2],"string":["A","B"]}`,
			setupContext:   false,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2],"string":["A","B"]}`,
		},
		{
			name:           "BindError",
			method:         http.MethodPatch,
			requestBody:    `{"number":[0],"string":["🇺🇸",]}`,
			setupContext:   false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset router for each test case
			s.SetupTest() // Assuming this resets the router

			// Create request
			var req *http.Request
			var err error
			if tc.requestBody != "" {
				req, err = http.NewRequest(tc.method, testJSONEndpoint, strings.NewReader(tc.requestBody))
			} else {
				req, err = http.NewRequest(tc.method, testJSONEndpoint, nil)
			}
			s.NoError(err)

			// Setup context if needed
			if tc.setupContext {
				s.router.Use(func(c *gin.Context) {
					inCtx := &testStruct{
						Number: []uint{1, 2, 3},
						String: []string{"A", "B", "C"},
					}
					c.Set(ctxKey, inCtx)
					c.Next()
				})
			}

			// Setup handler
			s.router.PATCH(testJSONEndpoint, shouldBindJSONHandler)

			// Run the test
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			// Verify results
			s.Equal(tc.expectedStatus, r.Code)

			if tc.verifyEmpty {
				// Special handling for empty struct comparison
				var empty testStruct
				bytes, err := json.Marshal(empty)
				s.NoError(err)
				s.ElementsMatch(bytes, r.Body.Bytes())
			} else if tc.expectedBody != "" {
				// Normal JSON comparison
				var expected, actual interface{}
				s.NoError(json.Unmarshal([]byte(tc.expectedBody), &expected))
				s.NoError(json.Unmarshal(r.Body.Bytes(), &actual))

				expectedJSON, err := json.Marshal(expected)
				s.NoError(err)
				actualJSON, err := json.Marshal(actual)
				s.NoError(err)

				s.Equal(string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func (s *BindingSuite) Test_BindJSON() {
	testCases := []struct {
		name           string
		method         string
		requestBody    string
		setupContext   bool
		contextValue   interface{}
		handler        gin.HandlerFunc
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "FromBody",
			method:         http.MethodPatch,
			requestBody:    `{"number":[1,2,3],"string":["A","B","C"]}`,
			setupContext:   false,
			handler:        bindJSONHandler,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"]}`,
		},
		{
			name:           "NoBody",
			method:         http.MethodPatch,
			requestBody:    "",
			setupContext:   false,
			handler:        bindJSONHandler,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"Op":"","Kind":400,"Err":null}`,
		},
		{
			name:           "FromBodyBindError",
			method:         http.MethodPatch,
			requestBody:    `{"number":[0],"string":["🇺🇸",]}`,
			setupContext:   false,
			handler:        bindJSONHandler,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:         "FromContext_OK",
			method:       http.MethodPatch,
			requestBody:  `{"number":[1,2,3],"string":["A","B","C"]}`,
			setupContext: true,
			contextValue: &testStruct{
				Number: []uint{1, 2, 3, 4},
				String: []string{"A", "B", "C", "D"},
			},
			handler:        bindJSONHandler,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3,4],"string":["A","B","C","D"]}`,
		},
		{
			name:           "FromContext_TypeMismatch",
			method:         http.MethodPatch,
			requestBody:    `{"number":[1,2,3],"string":["A","B","C"]}`,
			setupContext:   true,
			contextValue:   &struct{ Value string }{"test"}, // Different type than expected
			handler:        bindJSONHandler,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"]}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			var req *http.Request
			var err error
			if tc.requestBody != "" {
				req, err = http.NewRequest(tc.method, testJSONEndpoint, strings.NewReader(tc.requestBody))
			} else {
				req, err = http.NewRequest(tc.method, testJSONEndpoint, nil)
			}
			s.NoError(err)

			if tc.setupContext {
				s.router.Use(func(c *gin.Context) {
					c.Set(ctxKey, tc.contextValue)
					c.Next()
				})
			}

			s.router.PATCH(testJSONEndpoint, tc.handler)
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(tc.expectedStatus, r.Code)
			if tc.expectedBody != "" {
				// Normalize JSON for comparison
				var expected, actual interface{}
				err1 := json.Unmarshal([]byte(tc.expectedBody), &expected)
				err2 := json.Unmarshal(r.Body.Bytes(), &actual)

				if err1 == nil && err2 == nil {
					expectedJSON, _ := json.Marshal(expected)
					actualJSON, _ := json.Marshal(actual)
					s.Equal(string(expectedJSON), string(actualJSON))
				} else {
					// Direct string comparison as fallback
					s.Equal(tc.expectedBody, r.Body.String())
				}
			}
		})
	}
}

func (s *BindingSuite) Test_BindJSON_MultipleBinds() {
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "MultipleBinds_OK",
			requestBody: `{
				"number": [1,2,3],
				"string": ["A","B","C"],
				"dummy": "dummy"
			}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"],"dummy":"dummy"}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			req, err := http.NewRequest(http.MethodPatch, testMultipleBindEndpoint,
				strings.NewReader(tc.requestBody))
			s.NoError(err)

			verifyContext := func(c *gin.Context, ctxKey string, expected interface{}) {
				actual, exists := c.Get(ctxKey)
				s.True(exists)
				s.Equal(expected, actual)
			}

			verifyTestStructs := func(c *gin.Context) {
				inCtx := &testStruct{
					Number: []uint{1, 2, 3},
					String: []string{"A", "B", "C"},
				}
				anotherInCtx := &anotherTestStruct{
					Number: []uint{1, 2, 3},
					String: []string{"A", "B", "C"},
					Dummy:  "dummy",
				}
				verifyContext(c, ctxKey, inCtx)
				verifyContext(c, anotherCtxKey, anotherInCtx)
			}

			s.router.PATCH(testMultipleBindEndpoint, aBindJSONHandler, anotherBindJSONHandler, verifyTestStructs)
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(tc.expectedStatus, r.Code)

			if tc.expectedBody != "" {
				var expected, actual interface{}
				s.NoError(json.Unmarshal([]byte(tc.expectedBody), &expected))
				s.NoError(json.Unmarshal(r.Body.Bytes(), &actual))

				expectedJSON, _ := json.Marshal(expected)
				actualJSON, _ := json.Marshal(actual)

				s.Equal(string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func (s *BindingSuite) Test_BindQuery() {
	testCases := []struct {
		name           string
		params         url.Values
		setupContext   bool
		contextValue   interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "FromURL_BindError",
			params: func() url.Values {
				p := url.Values{}
				p.Add("Num", "0")
				p.Add("Str", "🇺🇸")
				return p
			}(),
			setupContext:   false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "FromURL_OK",
			params: func() url.Values {
				p := url.Values{}
				p.Add("Num", "1")
				p.Add("Num", "2")
				p.Add("Num", "3")
				p.Add("Str", "A")
				p.Add("Str", "B")
				p.Add("Str", "C")
				return p
			}(),
			setupContext:   false,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"]}`,
		},
		{
			name: "FromContext_OK",
			params: func() url.Values {
				p := url.Values{}
				p.Add("Num", "1")
				p.Add("Num", "2")
				p.Add("Num", "3")
				p.Add("Str", "A")
				p.Add("Str", "B")
				p.Add("Str", "C")
				return p
			}(),
			setupContext: true,
			contextValue: &testStruct{
				Number: []uint{1, 2, 3, 4},
				String: []string{"A", "B", "C", "D"},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3,4],"string":["A","B","C","D"]}`,
		},
		{
			name: "FromContext_TypeMismatch",
			params: func() url.Values {
				p := url.Values{}
				p.Add("Num", "1")
				p.Add("Num", "2")
				p.Add("Num", "3")
				p.Add("Str", "A")
				p.Add("Str", "B")
				p.Add("Str", "C")
				return p
			}(),
			setupContext:   true,
			contextValue:   &struct{ Value string }{"test"}, // Different type
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"]}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			req, err := http.NewRequest(http.MethodPost,
				fmt.Sprintf("%v?%v", testQueryEndpoint, tc.params.Encode()), nil)
			s.NoError(err)

			if tc.setupContext {
				s.router.Use(func(c *gin.Context) {
					c.Set(ctxKey, tc.contextValue)
					c.Next()
				})
			}

			s.router.POST(testQueryEndpoint, bindQueryHandler)
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(tc.expectedStatus, r.Code)

			if tc.expectedBody != "" {
				var expected, actual interface{}
				s.NoError(json.Unmarshal([]byte(tc.expectedBody), &expected))
				s.NoError(json.Unmarshal(r.Body.Bytes(), &actual))

				expectedJSON, _ := json.Marshal(expected)
				actualJSON, _ := json.Marshal(actual)

				s.Equal(string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func (s *BindingSuite) Test_BindChangeRequest() {
	testCases := []struct {
		name           string
		requestBody    string
		idParam        string
		setupContext   bool
		contextValue   interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "FromBody_BindError",
			requestBody: `{
				"number": [0],
				"string": ["🇺🇸",]
			}`,
			idParam:        "1",
			setupContext:   false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "FromBody_BindIDError",
			requestBody: `{
				"number": [1],
				"string": ["A"],
				"requestedBy": "admin@payments",
				"reason": "update"
			}`,
			idParam:        "0",
			setupContext:   false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "FromBody_OK",
			requestBody: `{
				"number": [1,2,3],
				"string": ["A","B","C"],
				"requestedBy": "admin@payments",
				"reason": "update"
			}`,
			idParam:        "1",
			setupContext:   false,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"],"requestedBy":"admin@payments","reason":"update"}`,
		},
		{
			name: "FromContext_TypeMismatch",
			requestBody: `{
				"number": [1,2,3],
				"string": ["A","B","C"],
				"requestedBy": "admin@payments",
				"reason": "update"
			}`,
			idParam:      "1",
			setupContext: true,
			contextValue: &testStruct{
				Number: []uint{1, 2, 3, 4},
				String: []string{"A", "B", "C", "D"},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3],"string":["A","B","C"],"requestedBy":"admin@payments","reason":"update"}`,
		},
		{
			name: "FromContext_OK",
			requestBody: `{
				"number": [1,2,3],
				"string": ["A","B","C"],
				"requestedBy": "admin@payments",
				"reason": "update"
			}`,
			idParam:      "1",
			setupContext: true,
			contextValue: &testChangeStruct{
				Number: []uint{1, 2, 3, 4},
				String: []string{"A", "B", "C", "D"},
				ChangeRequest: audit.ChangeRequest{
					ID: 2,
					Request: audit.Request{
						RequestedBy: pointer("admin@payments"),
						Reason:      pointer("update"),
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"number":[1,2,3,4],"string":["A","B","C","D"],"requestedBy":"admin@payments","reason":"update"}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			endpoint := testChangeRequestEndpoint + "/" + tc.idParam
			req, err := http.NewRequest(http.MethodPatch, endpoint, strings.NewReader(tc.requestBody))
			s.NoError(err)

			if tc.setupContext {
				s.router.Use(func(c *gin.Context) {
					c.Set(ctxKey, tc.contextValue)
					c.Next()
				})
			}

			s.router.PATCH(testChangeRequestEndpoint+"/:id", bindChangeRequestHandler)
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(tc.expectedStatus, r.Code)

			if tc.expectedStatus == http.StatusOK && tc.expectedBody != "" {
				var expected, actual interface{}
				s.NoError(json.Unmarshal([]byte(tc.expectedBody), &expected))
				s.NoError(json.Unmarshal(r.Body.Bytes(), &actual))

				expectedJSON, _ := json.Marshal(expected)
				actualJSON, _ := json.Marshal(actual)

				s.Equal(string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func (s *BindingSuite) Test_BindID() {
	testCases := []struct {
		name           string
		idParam        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "BadRequest",
			idParam:        "ABC",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "BadRequest_ZeroID",
			idParam:        "0",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "OK",
			idParam:        "1",
			expectedStatus: http.StatusOK,
			expectedBody:   "1",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/"+tc.idParam, nil)
			s.NoError(err)

			s.router.GET(testIDEndpoint+"/:id", bindIDHandler)
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(tc.expectedStatus, r.Code)
			if tc.expectedBody != "" {
				s.Equal(tc.expectedBody, r.Body.String())
			}
		})
	}
}

func (s *BindingSuite) Test_BindURIUint() {
	uriParam := "someUint"
	s.router.GET(testIDEndpoint+"/:"+uriParam, func(c *gin.Context) {
		id, err := binding.BindURIUint(c, uriParam)
		if err != nil {
			c.AbortWithStatusJSON(errors.Code(err), err)
			return
		}
		c.JSON(http.StatusOK, id)
	})

	// OK
	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/1", nil)
	s.NoError(err)

	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("1", r.Body.String())

	// Bad Request
	req, err = http.NewRequest(http.MethodGet, testIDEndpoint+"/ABC", nil)
	s.NoError(err)

	r = httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)

	// Zero ID
	req, err = http.NewRequest(http.MethodGet, testIDEndpoint+"/0", nil)
	s.NoError(err)

	r = httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindURI() {
	testCases := []struct {
		name           string
		id             string
		childID        string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "WithBody",
			id:             "123",
			childID:        "456",
			requestBody:    `{"name": "test name"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"123","child_id":"456","name":"test name"}`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			path := fmt.Sprintf("/test/%s/child/%s", tc.id, tc.childID)
			req, err := http.NewRequest(http.MethodPatch, path, strings.NewReader(tc.requestBody))
			s.NoError(err)

			s.router.PATCH(testUriEndpoint, bindUriHandler)
			r := httptest.NewRecorder()
			s.router.ServeHTTP(r, req)

			s.Equal(tc.expectedStatus, r.Code)

			if tc.expectedBody != "" {
				var expected, actual interface{}
				s.NoError(json.Unmarshal([]byte(tc.expectedBody), &expected))
				s.NoError(json.Unmarshal(r.Body.Bytes(), &actual))

				expectedJSON, _ := json.Marshal(expected)
				actualJSON, _ := json.Marshal(actual)

				s.Equal(string(expectedJSON), string(actualJSON))
			}
		})
	}
}
