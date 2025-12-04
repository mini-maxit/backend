package schemas

import "github.com/mini-maxit/backend/package/domain/types"

// AddCollaborator is the request schema for adding a collaborator to a contest or task.
type AddCollaborator struct {
	UserID     int64            `json:"user_id" validate:"required,min=1"`
	Permission types.Permission `json:"permission" validate:"required,oneof=edit manage"`
}

// UpdateCollaborator is the request schema for updating a collaborator's permission.
type UpdateCollaborator struct {
	Permission types.Permission `json:"permission" validate:"required,oneof=edit manage"`
}

// Collaborator represents a collaborator with their details.
type Collaborator struct {
	UserID     int64            `json:"userId"`
	UserName   string           `json:"userName"`
	FirstName  string           `json:"firstName"`
	LastName   string           `json:"lastName"`
	UserEmail  string           `json:"userEmail"`
	Permission types.Permission `json:"permission"`
	AddedAt    string           `json:"addedAt"`
}
