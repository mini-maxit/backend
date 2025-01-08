package schemas

type User struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UserEdit struct {
	Name     *string `json:"name,omitempty"`
	Surname  *string `json:"surname,omitempty"`
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
}
