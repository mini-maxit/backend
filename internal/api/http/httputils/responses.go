package httputils

type MessageResponse struct {
	Message string `json:"message"`
}

func NewMessageResponse(message string) *MessageResponse {
	return &MessageResponse{Message: message}
}

type IDResponse struct {
	ID int64 `json:"id"`
}

func NewIDResponse(id int64) *IDResponse {
	return &IDResponse{ID: id}
}
