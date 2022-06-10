package auth

import "gorm.io/gorm"

// Role ...
type Role struct {
	gorm.Model
	Name  string
	Users []*User `gorm:"many2many:auth_user_roles"`
}

// TableName return the table name
func (r *Role) TableName() string {
	return "auth_roles"
}

// User ...
type User struct {
	gorm.Model
	Email string
}

// TableName return the table name
func (r *User) TableName() string {
	return "auth_users"
}

// UserRoles ...
type UserRoles struct {
	gorm.Model
	RoleID uint
	UserID uint
}

// TableName return the table name
func (r *UserRoles) TableName() string {
	return "auth_user_roles"
}

// Permission ...
type Permission struct {
	gorm.Model
	RoleID   uint
	Role     *Role `gorm:"foreignKey:RoleID"`
	Resource string
	Method   string
}

// RoleName get role name of current permission
func (p *Permission) RoleName() (name string) {
	if p != nil && p.Role != nil {
		name = p.Role.Name
	}
	return
}

// TableName return the table name
func (p *Permission) TableName() string {
	return "auth_role_permissions"
}
