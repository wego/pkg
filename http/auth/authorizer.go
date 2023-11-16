package auth

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/wego/pkg/errors"
	"gorm.io/gorm"
)

var (
	actionMappings = map[string]string{
		http.MethodGet:    "Read",
		http.MethodPost:   "Create",
		http.MethodPut:    "Update",
		http.MethodPatch:  "Update",
		http.MethodDelete: "Delete",
	}
)

// Authorizer an RBAC authorizer
type Authorizer struct {
	enforcer    *casbin.Enforcer
	userHandler func(r *http.Request) (string, error)
}

// NewAuthorizer returns the authorizer
func NewAuthorizer(conf string, db *gorm.DB, userHandler func(r *http.Request) (string, error)) (*Authorizer, error) {
	adapter := newAdapter(db)
	e, err := casbin.NewEnforcer(conf, adapter)
	if err != nil {
		return nil, errors.New(nil, "can not create a Casbin Enforcer", err)
	}

	return &Authorizer{
		enforcer:    e,
		userHandler: userHandler,
	}, nil
}

// Auth returns the authorizer handler
func (a *Authorizer) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := a.checkPermission(c.Request); err != nil {
			c.AbortWithStatusJSON(errors.Code(err), gin.H{"errors": []string{err.Error()}})
		}
	}
}

// LoadPolicy load the policy from the adaptor
func (a *Authorizer) LoadPolicy() error {
	return a.enforcer.LoadPolicy()
}

// checkPermission checks the user/path/method combination from the request.
// Returns nil (permission granted) or error (permission denied)
func (a *Authorizer) checkPermission(r *http.Request) error {
	user, err := a.userHandler(r)
	if err != nil {
		return err
	}

	method := r.Method
	path := r.URL.Path
	allowed, err := a.enforcer.Enforce(user, path, method)
	if err != nil {
		// directly panic to throw errors, gin will recover the panic
		panic(err)
	}

	if !allowed {
		return errors.New(nil, errors.Forbidden, fmt.Sprintf("user %s is not allow to %s on %s", user, mappingAction(method), path))
	}

	return nil
}

func mappingAction(method string) string {
	if action, ok := actionMappings[method]; ok {
		return action
	}
	return method
}
