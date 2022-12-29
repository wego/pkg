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
	jwksURL   string
	jwtHeader string
)

// Init initializes the package with
//
//   - JWKS URL for getting JWKS. Format: https://www.rfc-editor.org/rfc/rfc7517#section-5)
//   - JWT header name for reading the JWT token from
//   - refresh interval for reloading the JWK cache
func Init(url, headerName string, refreshInterval time.Duration) error {
	jwksURL = url
	jwtHeader = headerName

	ctx := context.Background()
	jwkCache = jwk.NewCache(ctx)

	jwkCache.Register(jwksURL, jwk.WithMinRefreshInterval(refreshInterval))
	if _, err := jwkCache.Refresh(ctx, jwksURL); err != nil {
		return err
	}
	return nil
}

// GetJWTToken verify and return jwt token from http request, only accept bearer header.
//
// Make sure you call Init before can use this.
func GetJWTToken(req *http.Request) (jwt.Token, error) {
	if jwkCache == nil || jwksURL == "" || jwtHeader == "" {
		return nil, errors.New(errors.Unauthorized, "jwk cache has not been initialized")
	}

	authHeader := req.Header.Get(jwtHeader)
	if !strings.HasPrefix(authHeader, header.BearerPrefix) {
		return nil, errors.New(errors.Unauthorized, fmt.Sprintf("invalid %s header", jwtHeader))
	}

	token, err := jwt.Parse(
		[]byte(strings.TrimPrefix(authHeader, header.BearerPrefix)),
		jwt.WithKeySet(jwk.NewCachedSet(jwkCache, jwksURL)),
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
