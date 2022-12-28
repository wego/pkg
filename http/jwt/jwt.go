// Package jwt handles some business related to JWT token.
package jwt

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/wego/pkg/errors"
	"github.com/wego/pkg/http/header"
)

var (
	jwkCache  *jwk.Cache
	jwkURL    string
	jwtHeader string
)

// Init initializes the package with a JWK URL, a JWT header name & refresh interval for the JWK cache
func Init(url, headerName string, refreshInterval time.Duration) error {
	jwkURL = url
	jwtHeader = headerName

	ctx := context.Background()
	jwkCache = jwk.NewCache(ctx)

	jwkCache.Register(jwkURL, jwk.WithMinRefreshInterval(refreshInterval))
	if _, err := jwkCache.Refresh(ctx, jwkURL); err != nil {
		return err
	}
	return nil
}

// GetJWTToken verify and return jwt token from http request, only accept bearer header.
//
// Make sure you call Init before can use this.
func GetJWTToken(req *http.Request) (jwt.Token, error) {
	if jwkCache == nil || jwkURL == "" || jwtHeader == "" {
		return nil, errors.New(errors.Unauthorized, "jwk cache has not been initialized")
	}

	authHeader := req.Header.Get(jwtHeader)
	if !strings.HasPrefix(authHeader, header.BearerPrefix) {
		return nil, errors.New(errors.Unauthorized, fmt.Sprintf("invalid %s header", jwtHeader))
	}

	token, err := jwt.Parse(
		[]byte(strings.TrimPrefix(authHeader, header.BearerPrefix)),
		jwt.WithKeySet(jwk.NewCachedSet(jwkCache, jwkURL)),
	)
	if err != nil {
		return nil, errors.New(errors.Unauthorized, fmt.Sprintf("invalid jwt token: %s", err))
	}

	return token, err
}

// GetUserEmail return email private claim from jwt token in the http request
func GetUserEmail(req *http.Request) (email string, err error) {
	token, err := GetJWTToken(req)
	if err != nil {
		return
	}

	return token.PrivateClaims()["email"].(string), nil
}
