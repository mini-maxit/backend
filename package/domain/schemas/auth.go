package schemas

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
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
