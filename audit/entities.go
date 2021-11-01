package audit

// Entity base fields used in audit
type Entity struct {
	LastUpdatedBy *string     `gorm:"-"`
	UpdateReason  *string     `gorm:"-"`
	ActionType    *ActionType `gorm:"-"`
}

// ActionEntity base fields used in actions tables
type ActionEntity struct {
	UpdatedBy    *string
	UpdateReason *string `gorm:"column:update_reason"`
	ActionType   ActionType
}
