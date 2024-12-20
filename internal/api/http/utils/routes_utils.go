package utils

import (
	"encoding/json"
	"net/http"
)

type ApiResponse[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

const DefaultPaginationLimitStr = "10"
const DefaultPaginationOffsetStr = "0"

const (
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ReturnError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ApiResponse[any]{
		Ok:   false,
		Data: Error{Code: code, Message: message},
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, CodeInternalServerError, err.Error())
		return
	}
}

func ReturnSuccess(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ApiResponse[any]{
		Ok:   true,
		Data: data,
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, CodeInternalServerError, err.Error())
		return
	}
}
