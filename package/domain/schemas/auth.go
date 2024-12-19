package schemas

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Name     string `json:"name" validate:"nonzero,min=3,max=50"`
	Surname  string `json:"surname" validate:"nonzero,min=3,max=50"`
	Email    string `json:"email" validate:"nonzero,regexp=^(?:(?!\.)[\w\-_.]*[^.])(?:@\w+)(?:\.\w+(?:\.\w+)?[^.\W])$"`
	Username string `json:"username" validate:"nonzero,min=3,max=30,regexp=^[a-zA-Z][a-zA-Z0-9_]*[a-zA-Z0-9]$"`
	Password string `json:"password" validate:"nonzero,min=8,max=50"`
}
