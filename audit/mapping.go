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
// When we create a model with audit like
//  type XXXModel struct {
//	   ... // custom business fields
//	   Actions []*audit.Action
//     audit.Model
// }
// And the corresponding DB XXXEntity
//  type XXXModel struct {
//     gorm.Model
//	   ... // custom business fields
//	   Actions []*audit.ActionEntity
//     audit.Entity
// }
// When we want to map from the XXXEntity to XXXModel
// just do as below
// func XXXEntityToXXXModel(e *XXXEntity) * XXXModel {
//    return &XXXModel{
//        ... // mapping custom business fields
//        Model: EntityToModel(&e.Entity, &e.Model)
//    }
// }
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
// When we create a model with audit action like
//  type XXXAction struct {
//	   ... // custom business fields
//     audit.Action
// }
// And the corresponding DB XXXActionEntity
//  type XXXActionEntity struct {
//     gorm.Model
//	   ... // custom business fields
//     audit.ActionEntity
// }
// When we want to map from the XXXEntity to XXXModel
// just do as below
// func XXXActionEntityToXXXAction(e *XXXActionEntity) * XXXAction {
//    return &XXXAction{
//        ... // mapping custom business fields
//        ActionEntity: ActionEntityToAction(&e.Entity, &e.Model)
//    }
// }
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
