package schemas

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserRegisterRequest struct {
	Name            string `json:"name" validate:"required,gte=3,lte=50"`
	Surname         string `json:"surname" validate:"required,gte=3,lte=50"`
	Email           string `json:"email" validate:"required,email"`
	Username        string `json:"username" validate:"required,gte=3,lte=30,username"`
	Password        string `json:"password" validate:"required,gte=8,lte=50,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,password,eqfield=Password"`
}
