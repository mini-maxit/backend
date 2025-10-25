package schemas

import "time"

type GroupDetailed struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Tasks     []Task    `json:"tasks"`
	Users     []User    `json:"users"`
}

type Group struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateGroup struct {
	Name string `json:"name" validate:"required,gte=3,lte=50"`
}

type EditGroup struct {
	Name *string `json:"name,omitempty" validate:"gte=3,lte=50"`
}
