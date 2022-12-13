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
	"github.com/wego/pkg/common"
	"github.com/wego/pkg/errors"
	"github.com/wego/pkg/http/binding"
)

var (
	testJSONEndpoint          = "/test/json"
	testQueryEndpoint         = "/test/query"
	testChangeRequestEndpoint = "/test/cr"
	testIDEndpoint            = "/test"
	testUriEndpoint           = "/test/:id/child/:child_id"
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

type testUriStruct struct {
	ID      string `uri:"id" json:"id"`
	ChildID string `uri:"child_id" json:"child_id"`
	Name    string `json:"name"`
}

func shouldBindJSONHandler(c *gin.Context) {
	var t testStruct
	if err := binding.ShouldBindJSON(c, ctxKey, &t); err != nil {
		c.AbortWithStatus(errors.Code(err))
		return
	}
	c.JSON(http.StatusOK, t)
}

func bindJSONHandler(c *gin.Context) {
	var t testStruct
	if err := binding.BindJSON(c, ctxKey, &t); err != nil {
		c.AbortWithStatus(errors.Code(err))
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
		c.AbortWithStatus(errors.Code(err))
		return
	}
	c.JSON(http.StatusOK, t)
}

func bindIDHandler(c *gin.Context) {
	id, err := binding.BindID(c)
	if err != nil {
		c.AbortWithStatus(errors.Code(err))
		return
	}
	c.JSON(http.StatusOK, id)
}

func bindUriHandler(c *gin.Context) {
	var t testUriStruct
	ctxKeyUri := "keyUri"
	ctxKeyBody := "keyBody"

	if err := binding.BindUri(c, ctxKeyUri, &t); err != nil {
		c.AbortWithStatus(errors.Code(err))
		return
	}
	if err := binding.BindJSON(c, ctxKeyBody, &t); err != nil {
		c.AbortWithStatus(errors.Code(err))
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

func (s *BindingSuite) Test_ShouldBindJSON_FromContext() {
	requestBody := `{
	    "number": [
	        1
	    ],
	    "string": [
	        "A"
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	inCtx := &testStruct{
		Number: []uint{1, 2, 3},
		String: []string{"A", "B", "C"},
	}
	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, inCtx)
		c.Next()
	})
	s.router.PATCH(testJSONEndpoint, shouldBindJSONHandler)

	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}", r.Body.String())
}

func (s *BindingSuite) Test_ShouldBindJSON_NoBody() {
	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, nil)
	s.NoError(err)

	s.router.PATCH(testJSONEndpoint, shouldBindJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	var empty testStruct
	bytes, err := json.Marshal(empty)
	s.NoError(err)
	s.ElementsMatch(bytes, r.Body.Bytes())
}

func (s *BindingSuite) Test_ShouldBindJSON_FromBody() {
	requestBody := `{
	    "number": [
	        1,
	        2
	    ],
	    "string": [
	        "A",
	        "B"
	    ]
	}`
	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testJSONEndpoint, shouldBindJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2],\"string\":[\"A\",\"B\"]}", r.Body.String())
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

	s.router.PATCH(testJSONEndpoint, bindJSONHandler)
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

	s.router.PATCH(testJSONEndpoint, bindJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}", r.Body.String())
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

	inCtx := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}
	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, inCtx)
		c.Next()
	})
	s.router.PATCH(testJSONEndpoint, bindJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	body, err := json.Marshal(&inCtx)
	s.NoError(err)
	s.Equal(string(body), r.Body.String())
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
	inCtx := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &inCtx)
		c.Next()
	})

	req, err := http.NewRequest(http.MethodPatch, testJSONEndpoint, strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testJSONEndpoint, bindJSONHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}", r.Body.String())
}

func (s *BindingSuite) Test_BindQuery_FromURL_BindError() {
	s.router.POST(testQueryEndpoint, bindQueryHandler)
	params := url.Values{}
	params.Add("Num", "0")
	params.Add("Str", "🇺🇸")
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v?%v", testQueryEndpoint, params.Encode()), nil)
	s.NoError(err)

	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindQuery_FromURL_OK() {
	params := url.Values{}
	params.Add("Num", "1")
	params.Add("Num", "2")
	params.Add("Num", "3")
	params.Add("Str", "A")
	params.Add("Str", "B")
	params.Add("Str", "C")
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v?%v", testQueryEndpoint, params.Encode()), nil)
	s.NoError(err)

	s.router.POST(testQueryEndpoint, bindQueryHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}", r.Body.String())
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

	inCtx := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, inCtx)
		c.Next()
	})
	s.router.POST(testQueryEndpoint, bindQueryHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	body, err := json.Marshal(&inCtx)
	s.NoError(err)
	s.Equal(string(body), r.Body.String())
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

	inCtx := &testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}

	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &inCtx)
		c.Next()
	})
	s.router.POST(testQueryEndpoint, bindQueryHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)

	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}", r.Body.String())
}

func (s *BindingSuite) Test_BindChangeRequest_FromBody_BindError() {
	s.router.PATCH(testChangeRequestEndpoint+"/:id", bindChangeRequestHandler)
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
	        "A"
	    ],
		"requestedBy": "admin@payments",
		"reason": "update"
	}`
	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/0", strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testChangeRequestEndpoint+"/:id", bindChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
	s.Empty(r.Body)
}

func (s *BindingSuite) Test_BindChangeRequest_FromBody_OK() {
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

	s.router.PATCH(testChangeRequestEndpoint+"/:id", bindChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"],\"requestedBy\":\"admin@payments\",\"reason\":\"update\"}", r.Body.String())
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

	inCtx := testStruct{
		Number: []uint{1, 2, 3, 4},
		String: []string{"A", "B", "C", "D"},
	}
	s.router.Use(func(c *gin.Context) {
		c.Set(ctxKey, &inCtx)
		c.Next()
	})
	s.router.PATCH(testChangeRequestEndpoint+"/:id", bindChangeRequestHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"],\"requestedBy\":\"admin@payments\",\"reason\":\"update\"}", r.Body.String())
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

	inCtx := testChangeStruct{
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
		c.Set(ctxKey, &inCtx)
		c.Next()
	})
	s.router.PATCH(testChangeRequestEndpoint+"/:id", bindChangeRequestHandler)
	req, err := http.NewRequest(http.MethodPatch, testChangeRequestEndpoint+"/1", strings.NewReader(requestBody))
	s.NoError(err)

	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	body, err := json.Marshal(inCtx)
	s.NoError(err)
	s.Equal(string(body), r.Body.String())
}

func (s *BindingSuite) Test_BindID_BadRequest() {
	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/ABC", nil)
	s.NoError(err)

	s.router.GET(testIDEndpoint+"/:id", bindIDHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindID_BadRequest_ZeroID() {

	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/0", nil)
	s.NoError(err)

	s.router.GET(testIDEndpoint+"/:id", bindIDHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusBadRequest, r.Code)
}

func (s *BindingSuite) Test_BindID_OK() {
	req, err := http.NewRequest(http.MethodGet, testIDEndpoint+"/1", nil)
	s.NoError(err)

	s.router.GET(testIDEndpoint+"/:id", bindIDHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("1", r.Body.String())
}

func (s *BindingSuite) Test_BindURIUint() {
	uriParam := "someUint"
	s.router.GET(testIDEndpoint+"/:"+uriParam, func(c *gin.Context) {
		id, err := binding.BindURIUint(c, uriParam)
		if err != nil {
			c.AbortWithStatus(errors.Code(err))
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

func (s *BindingSuite) Test_BindUri_WithBody() {
	requestBody := `{
	    "name": "test name"
	}`
	req, err := http.NewRequest(http.MethodPatch, "/test/123/child/456", strings.NewReader(requestBody))
	s.NoError(err)

	s.router.PATCH(testUriEndpoint, bindUriHandler)
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)

	s.Equal(http.StatusOK, r.Code)
	s.Equal("{\"id\":\"123\",\"child_id\":\"456\",\"name\":\"test name\"}", r.Body.String())
}
