package middleware

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"strings"
)

// QueryArraySupport since gin does not support comma seperated array like `?&query=1,2,3`, this middle try to do a workaround
// Use this with cautions, this will break the normal query parameter containing comma
func QueryArraySupport() gin.HandlerFunc {
	return func(c *gin.Context) {
		updated := url.Values{}
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if len(value) == 0 {
					continue
				}
				for _, split := range strings.Split(value, ",") {
					if len(updated.Get(key)) == 0 {
						updated.Set(key, split)
					} else {
						updated.Add(key, split)
					}
				}
			}
		}
		c.Request.URL.RawQuery = updated.Encode()
		c.Next()
	}
}
