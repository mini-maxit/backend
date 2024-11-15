package utils

import (
	"encoding/json"
	"net/http"
)

type ApiResponse[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

func ReturnError(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ApiResponse[any]{
		Ok:   false,
		Data: data,
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
}

func ReturnInternalServerError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
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
		ReturnInternalServerError(w, err)
		return
	}
}
