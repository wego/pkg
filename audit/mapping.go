package admin

import (
	"gorm.io/gorm"
)

// RequestToEntity request to entity
func RequestToEntity(req *Request) (entity Entity) {
	if req != nil {
		entity.LastUpdatedBy = req.RequestedBy
		entity.UpdateReason = req.Reason
	}
	return
}

// EntityToModel base entity to base model
func EntityToModel(entity *Entity, m *gorm.Model) (model Model) {
	if entity != nil {
		model.UpdatedBy = entity.LastUpdatedBy
		model.UpdateReason = entity.UpdateReason
	}
	if m != nil {
		if !m.CreatedAt.IsZero() {
			model.CreatedAt = &m.CreatedAt
		}
		if !m.UpdatedAt.IsZero() {
			model.UpdatedAt = &m.UpdatedAt
		}
	}
	return
}

// ActionEntityToAction action entity to action
func ActionEntityToAction(entity *ActionEntity, model *gorm.Model) (action Action) {
	if entity != nil {
		action.ActionType = entity.ActionType
		action.UpdatedBy = entity.UpdatedBy
		action.UpdateReason = entity.UpdateReason
	}

	if model != nil {
		if !model.CreatedAt.IsZero() {
			action.CreatedAt = &model.CreatedAt
		}
		if !model.UpdatedAt.IsZero() {
			action.UpdatedAt = &model.UpdatedAt
		}
	}

	return
}
