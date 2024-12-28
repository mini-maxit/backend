package schemas

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=8,lte=50"`
}

type UserRegisterRequest struct {
	Name     string `json:"name" validate:"required,gte=3,lte=50"`
	Surname  string `json:"surname" validate:"required,gte=3,lte=50"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,gte=3,lte=30,username"`
	Password string `json:"password" validate:"required,gte=8,lte=50"`
}

// Structures defining the response of the API
type UserLoginSuccessResponse struct {
	Token string `json:"token"`
}

type UserLoginErrorResponse struct {
	Message string `json:"message"`
}

type UserRegisterSuccessResponse struct {
	SessionID string `json:"session_id"`
}

type UserRegisterErrorResponse struct {
	Message string `json:"message"`
}
