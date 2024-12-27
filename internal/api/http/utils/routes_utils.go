package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ApiResponse[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ReturnError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	code := http.StatusText(statusCode)
	response := ApiResponse[any]{
		Ok:   false,
		Data: Error{Code: code, Message: message},
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, err.Error())
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
		ReturnError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func GetLimitAndOffset(limitStr string, offsetStr string) (int64, int64, error) {
	if limitStr == "" {
		limitStr = DefaultPaginationLimitStr
	}
	if offsetStr == "" {
		offsetStr = DefaultPaginationOffsetStr
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return limit, offset, nil
}
