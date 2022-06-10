package auth

import (
	"errors"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

const (
	permissionTag = "p"
	groupTag      = "g"
)

// adapter is a customized read-only gorm adapter for `Casbin`.
// It can load policy from db
type adapter struct {
	db *gorm.DB
}

func (a *adapter) UpdatePolicy(string, string, []string, []string) error {
	return errors.New("not implemented")
}

func (a *adapter) UpdatePolicies(string, string, [][]string, [][]string) error {
	return errors.New("not implemented")
}

func (a *adapter) UpdateFilteredPolicies(string, string, [][]string, int, ...string) ([][]string, error) {
	return nil, errors.New("not implemented")
}

// newAdapter is the constructor for adapter.
func newAdapter(db *gorm.DB) *adapter {
	if db == nil {
		panic("db is nil")
	}

	sqlDB, err := db.DB()

	if err != nil {
		panic("can not get db connection")
	}

	if err = sqlDB.Ping(); err != nil {
		panic(err)
	}

	if err = db.SetupJoinTable(&Role{}, "Users", &UserRoles{}); err != nil {
		panic(err)
	}
	return &adapter{db: db}
}

// LoadPolicy loads all policy rules from the storage.
func (a *adapter) LoadPolicy(model model.Model) error {
	var permissions []*Permission
	if err := a.db.Preload("Role").Order("role_id").Find(&permissions).Error; err != nil {
		return err
	}

	for _, permission := range permissions {
		loadPolicyLine(permission, model)
	}

	var roles []*Role
	if err := a.db.Preload("Users").Find(&roles).Error; err != nil {
		return err
	}

	for _, role := range roles {
		loadRoleLine(role, model)
	}

	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *adapter) SavePolicy(model model.Model) error {
	return errors.New("not implemented")
}

func loadPolicyLine(permission *Permission, model model.Model) {
	if permission == nil {
		return
	}
	p := []string{permissionTag, permission.RoleName(), permission.Resource, permission.Method}
	persist.LoadPolicyArray(p, model)
}

func loadRoleLine(role *Role, model model.Model) {
	if role == nil || len(role.Users) == 0 {
		return
	}
	for _, user := range role.Users {
		if user == nil {
			continue
		}

		g := []string{groupTag, user.Email, role.Name}
		persist.LoadPolicyArray(g, model)
	}
}

// AddPolicy adds a policy rule to the storage.
func (a *adapter) AddPolicy(string, string, []string) error {
	return errors.New("not implemented")
}

// AddPolicies adds policy rules to the storage.
func (a *adapter) AddPolicies(string, string, [][]string) error {
	return errors.New("not implemented")
}

// RemovePolicy removes a policy rule from the storage.
func (a *adapter) RemovePolicy(string, string, []string) error {
	return errors.New("not implemented")
}

// RemovePolicies removes policy rules from the storage.
func (a *adapter) RemovePolicies(string, string, [][]string) error {
	return errors.New("not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *adapter) RemoveFilteredPolicy(string, string, int, ...string) error {
	return errors.New("not implemented")
}
