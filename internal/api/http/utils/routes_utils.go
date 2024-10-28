package utils

import (
	"encoding/json"
	"net/http"
)

type ApiResponse[T any] struct {
	ResponseKey string `json:"response_key"`
	Message     T      `json:"message"`
}

const (
	SuccessMessage   = "success"
	NotFound         = "not found"
	MethodNotAllowed = "method not allowed"
	BadRequest       = "bad request"
)

var (
	apiMessage = map[int]string{
		http.StatusOK:               SuccessMessage,
		http.StatusMethodNotAllowed: MethodNotAllowed,
		http.StatusNotFound:         NotFound,
		http.StatusBadRequest:       BadRequest,
	}
)

func ReturnError(w http.ResponseWriter, statusCode int, message any) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	response := ApiResponse[any]{
		ResponseKey: apiMessage[statusCode],
		Message:     message,
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
}

func ReturnInternalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(err.Error()))
}

func ReturnSuccess(w http.ResponseWriter, statusCode int, message any) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	response := ApiResponse[any]{
		ResponseKey: apiMessage[statusCode],
		Message:     message,
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
}
