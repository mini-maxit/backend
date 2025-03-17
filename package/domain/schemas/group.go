package schemas

import "time"

type GroupDetailed struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tasks     []Task    `json:"tasks"`
	Users     []User    `json:"users"`
}

type Group struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateGroup struct {
	Name string `json:"name" validate:"required,gte=3,lte=50"`
}

type EditGroup struct {
	Name *string `json:"name,omitempty" validate:"gte=3,lte=50"`
}
