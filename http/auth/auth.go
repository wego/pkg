package auth

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/wego/pkg/errors"
	"gorm.io/gorm"
	"net/http"
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
	userHandler func(r *http.Request) string
}

// NewAuthorizer returns the authorizer
func NewAuthorizer(conf string, db *gorm.DB, userHandler func(r *http.Request) string) (*Authorizer, error) {
	a, err := newAdapter(db)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(conf, a)
	if err != nil {
		return nil, errors.New("can not create a Casbin Enforcer", err)
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
			a.requirePermission(c, err)
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
	user := a.userHandler(r)
	method := r.Method
	path := r.URL.Path

	allowed, err := a.enforcer.Enforce(user, path, method)
	if err != nil {
		// directly panic to throw errors, gin will recover the panic
		panic(err)
	}

	if !allowed {
		return errors.New(fmt.Errorf("user %s is not allow to %s on %s", user, mappingAction(method), path))
	}

	return nil
}

// requirePermission returns the 403 Forbidden to the client
func (a *Authorizer) requirePermission(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"errors": []string{err.Error()}})
}

func mappingAction(method string) string {
	if action, ok := actionMappings[method]; ok {
		return action
	}
	return method
}
