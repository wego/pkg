package admin

import (
	"time"
)

// Request base request
type Request struct {
	// usually an email address, but can be something else
	// for email see the limit here https://stackoverflow.com/a/574698
	// for other like names or ids 254 is also enough
	RequestedBy *string `json:"requestedBy,omitempty"  binding:"required,printascii,max=254"`
	Reason      *string `json:"reason,omitempty"  binding:"required,printascii,max=512"`
}

// IChangeRequest general change request interface
type IChangeRequest interface {
	SetID(uint)
}

// ChangeRequest request to change a model
type ChangeRequest struct {
	ID uint `json:"-"`
	Request
}

// SetID set the model ID
func (r *ChangeRequest) SetID(id uint) {
	r.ID = id
}

// DeleteRequest request to delete a model
type DeleteRequest struct {
	ChangeRequest
}

// Model basic response fields
type Model struct {
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy    *string    `json:"updatedBy,omitempty"`
	UpdateReason *string    `json:"updateReason,omitempty"`
}

// Action base fields used in actions
type Action struct {
	ActionType   ActionType `json:"actionType,omitempty"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy    *string    `json:"updatedBy,omitempty"`
	UpdateReason *string    `json:"updateReason,omitempty"`
}

// After check one action is created after another
func (r Action) After(l Action) bool {
	cr := r.CreatedAt
	cl := l.CreatedAt
	if cr != nil && cl != nil {
		return cr.After(*cl)
	}

	if cl == nil && cr != nil {
		return !cr.IsZero()
	}
	return false
}

// Before check one action is created before another
func (r Action) Before(l Action) bool {
	cr := r.CreatedAt
	cl := l.CreatedAt
	if cr != nil && cl != nil {
		return cr.Before(*cl)
	}

	if cl != nil && cr == nil {
		return !cl.IsZero()
	}

	return false
}

// ActionType  action type
type ActionType string

// CRUD actions
const (
	Create ActionType = "Create"
	Update ActionType = "Update"
	Delete ActionType = "Delete"
)

// NewActionType create a new ActionType ref
func NewActionType(v ActionType) *ActionType {
	return &v
}

// ActionCreate ...
func ActionCreate() *ActionType {
	return NewActionType(Create)
}

// ActionUpdate ...
func ActionUpdate() *ActionType {
	return NewActionType(Update)
}

// ActionDelete ...
func ActionDelete() *ActionType {
	return NewActionType(Delete)
}
