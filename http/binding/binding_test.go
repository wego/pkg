package binding_test

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/audit"
	"github.com/wego/pkg/common"
	"github.com/wego/pkg/errors"
	"github.com/wego/pkg/http/binding"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var (
	testJSONEndpoint          = "/test/json"
	testQueryEndpoint         = "/test/query"
	testChangeRequestEndpoint = "/test/cr"
	testIDEndpoint            = "/test"
	ctxKey                    = "testRequest"
)

type testStruct struct {
	Number []uint   `form:"Num,omitempty" json:"number" binding:"required,dive,number,min=1"`
	String []string `form:"Str,omitempty" json:"string" binding:"required,dive,printascii"`
}

type testChangeStruct struct {
	Number []uint   `form:"Num,omitempty" json:"number" binding:"required,dive,number,min=1"`
	String []string `form:"Str,omitempty" json:"string" binding:"required,dive,printascii"`
	audit.ChangeRequest
}

func testJSONHandler(c *gin.Context) {
	var t testStruct
	if err := binding.BindJSON(c, ctxKey, &t); err != nil {
		c.AbortWithStatus(errors.Code(err))
	}
	c.JSON(http.StatusOK, t)
}

func testQueryHandler(c *gin.Context) {
	var t testStruct
	if err := binding.BindQuery(c, ctxKey, &t); err != nil {
		c.AbortWithStatusJSON(errors.Code(err), err)
	}
	c.JSON(http.StatusOK, t)
}

func testChangeRequestHandler(c *gin.Context) {
	var t testChangeStruct
	if err := binding.BindChangeRequest(c, ctxKey, &t); err != nil {
		c.AbortWithStatus(errors.Code(err))
	}
	c.JSON(http.StatusOK, t)
}

func testIDtHandler(c *gin.Context) {
	id, err := binding.BindID(c)
	if err != nil {
		c.AbortWithStatus(errors.Code(err))
	}
	c.JSON(http.StatusOK, id)
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

func (s *BindingSuite) Test_BindJSON_FromBodyBindError() {
	requestBody := `{
	    "number": [
	        0
	    ],
	    "string": [
	        "🇺🇸",
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testJSONEndpoint, testJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindJSON_FromBody() {
	requestBody := `{
	    "number": [
	        1,
	        2,
	        3
	    ],
	    "string": [
	        "A",
	        "B",
	        "C"
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testJSONEndpoint, testJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}")
}

func (s *BindingSuite) Test_BindJSON_FromContext_OK() {
	requestBody := `{
	    "number": [
	        1,
	        2,
	        3
	    ],
	    "string": [
	        "A",
	        "B",
	        "C"
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	testReq := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}
	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, testReq)
		c.Next()
	})
	s.router.PATCH(testJSONEndpoint, testJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	body, err := json.Marshal(&testReq)
	s.NoError(err)
	s.Equal(r.Body.String(), string(body))
}

func (s *BindingSuite) Test_BindJSON_FromContext_TypeMismatch() {
	requestBody := `{
	    "number": [
	        1,
	        2,
	        3
	    ],
	    "string": [
	        "A",
	        "B",
	        "C"
	    ]
	}`
	testReq := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &testReq)
		c.Next()
	})

	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testJSONEndpoint, testJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	s.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}")
}

func (s *BindingSuite) Test_BindQuery_FromBodyBindError() {
	s.router.POST(testQueryEndpoint, testQueryHandler)
	params := url.Values{}
	params.Add("Num", "0")
	params.Add("Str", "🇺🇸")
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v?%v", testQueryEndpoint, params.Encode()), nil)
	s.NoError(err)

	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindQuery_FromBody() {
	params := url.Values{}
	params.Add("Num", "1")
	params.Add("Num", "2")
	params.Add("Num", "3")
	params.Add("Str", "A")
	params.Add("Str", "B")
	params.Add("Str", "C")
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v?%v", testQueryEndpoint, params.Encode()), nil)
	s.NoError(err)

	s.router.POST(testQueryEndpoint, testQueryHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}")
}

func (s *BindingSuite) Test_BindQuery_FromContext_OK() {
	params := url.Values{}
	params.Add("Num", "1")
	params.Add("Num", "2")
	params.Add("Num", "3")
	params.Add("Str", "A")
	params.Add("Str", "B")
	params.Add("Str", "C")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v?%v", testQueryEndpoint, params.Encode()), nil)
	s.NoError(err)

	testReq := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, testReq)
		c.Next()
	})
	s.router.POST(testQueryEndpoint, testQueryHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	body, err := json.Marshal(&testReq)
	s.NoError(err)
	s.Equal(r.Body.String(), string(body))
}

func (s *BindingSuite) Test_BindQuery_FromContext_TypeMismatch() {
	params := url.Values{}
	params.Add("Num", "1")
	params.Add("Num", "2")
	params.Add("Num", "3")
	params.Add("Str", "A")
	params.Add("Str", "B")
	params.Add("Str", "C")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v?%v", testQueryEndpoint, params.Encode()), nil)
	s.NoError(err)

	testReq := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &testReq)
		c.Next()
	})
	s.router.POST(testQueryEndpoint, testQueryHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	s.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}")
}

func (s *BindingSuite) Test_BindChangeRequest_FromBody_BindError() {
	s.router.PATCH(testChangeRequestEndpoint+"/:id", testChangeRequestHandler)
	requestBody := `{
	    "number": [
	        0
	    ],
	    "string": [
	        "🇺🇸",
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/1", strings.NewReader(requestBody))
	s.NoError(err)

	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindChangeRequest_FromBody_BindIDError() {
	requestBody := `{
	    "number": [
	        1
	    ],
	    "string": [
	        "A",
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/0", strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testChangeRequestEndpoint+"/:id", testChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindChangeRequest_FromBody_Ok() {
	requestBody := `{
	    "number": [
	        1,
	        2,
	        3
	    ],
	    "string": [
	        "A",
	        "B",
	        "C"
	    ],
		"requestedBy": "admin@payments",
		"reason": "update"
	}`
	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/1", strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testChangeRequestEndpoint+"/:id", testChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"],\"requestedBy\":\"admin@payments\",\"reason\":\"update\"}")
}

func (s *BindingSuite) Test_BindChangeRequest_FromContext_TypeMismatch() {
	requestBody := `{
	    "number": [
	        1,
	        2,
	        3
	    ],
	    "string": [
	        "A",
	        "B",
	        "C"
	    ],
		"requestedBy": "admin@payments",
		"reason": "update"
	}`
	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/1", strings.NewReader(requestBody))
	s.NoError(err)

	testReq := testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}
	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &testReq)
		c.Next()
	})
	s.router.PATCH(testChangeRequestEndpoint+"/:id", testChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"],\"requestedBy\":\"admin@payments\",\"reason\":\"update\"}")
}

func (s *BindingSuite) Test_BindChangeRequest_FromContext_OK() {
	requestBody := `{
	    "number": [
	        1,
	        2,
	        3
	    ],
	    "string": [
	        "A",
	        "B",
	        "C"
	    ],
		"requestedBy": "admin@payments",
		"reason": "update"
	}`

	testReq := testChangeStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
		ChangeRequest: audit.ChangeRequest{
			ID: 2,
			Request: audit.Request{
				RequestedBy: common.StrRef("admin@payments"),
				Reason:      common.StrRef("update"),
			},
		},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &testReq)
		c.Next()
	})

	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/1", strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testChangeRequestEndpoint+"/:id", testChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	body, err := json.Marshal(testReq)
	s.NoError(err)
	s.Equal(r.Body.String(), string(body))
}

func (s *BindingSuite) Test_BindID_BadRequest() {

	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/ABC", nil)
	s.NoError(err)

	s.router.GET(testIDEndpoint+"/:id", testIDtHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindID_BadRequest_ZeroID() {

	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/0", nil)
	s.NoError(err)

	s.router.GET(testIDEndpoint+"/:id", testIDtHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindID_BadRequest_Ok() {

	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/1", nil)
	s.NoError(err)

	s.router.GET(testIDEndpoint+"/:id", testIDtHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal(r.Body.String(), "1")
}