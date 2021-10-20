package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/wego/pkg/http/middleware"
)

var (
	testEndpoint = "/test"
)

type testStruct struct {
	Number []uint   `form:"Num,omitempty" json:"number"`
	String []string `form:"Str,omitempty" json:"string"`
}

func testHandler(c *gin.Context) {
	var t testStruct
	if err := c.ShouldBindQuery(&t); err != nil {
		c.Status(http.StatusBadRequest)
	}
	c.JSON(http.StatusOK, t)
}

type MiddlewareSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupTest runs before each Test
func (m *MiddlewareSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	m.router = gin.Default()
	m.router.Use(middleware.QueryArraySupport())
	m.router.GET(testEndpoint, testHandler)
}

func TestHandlers(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}

func (m *MiddlewareSuite) Test_QueryArraySupport_Standard() {
	params := url.Values{}
	params.Add("Num", "1")
	params.Add("Num", "2")
	params.Add("Num", "3")
	params.Add("Str", "A")
	params.Add("Str", "B")
	params.Add("Str", "C")
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v?%v", testEndpoint, params.Encode()), nil)
	m.NoError(err)

	r := httptest.NewRecorder()
	m.router.ServeHTTP(r, req)

	m.Equal(http.StatusOK, r.Code)
	m.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}")
}

func (m *MiddlewareSuite) Test_QueryArraySupport_CommaSeperated() {
	params := url.Values{}
	params.Add("Num", "1,2,3")
	params.Add("Str", "A,B,C")
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v?%v", testEndpoint, params.Encode()), nil)
	m.NoError(err)

	r := httptest.NewRecorder()
	m.router.ServeHTTP(r, req)

	m.Equal(http.StatusOK, r.Code)
	m.Equal(r.Body.String(), "{\"number\":[1,2,3],\"string\":[\"A\",\"B\",\"C\"]}")
}
