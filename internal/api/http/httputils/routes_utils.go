package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ApiResponse[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

type errorStruct struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ApiError ApiResponse[errorStruct]

func ReturnError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	code := http.StatusText(statusCode)
	response := ApiError{
		Ok:   false,
		Data: errorStruct{Code: code, Message: message},
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

func SetDefaultQueryParams(query *url.Values, field string) {
	if query.Get("limit") == "" {
		query.Set("limit", DefaultPaginationLimitStr)
	}
	if query.Get("offset") == "" {
		query.Set("offset", DefaultPaginationOffsetStr)
	}
	if query.Get("sort") == "" {
		sort := fmt.Sprintf("%s:%s", field, DefaultSortOrder)
		query.Set("sort", sort)
	}
}

func GetQueryParams(query *url.Values, field string) (map[string]string, error) {
	queryParams := map[string]string{}
	SetDefaultQueryParams(query, field)
	for key, value := range *query {
		if(len(value) > 1) {
			err := QueryError{Filed: key, Detail: MultipleQueryValues}
			return nil, err
		}
		queryParams[key] = value[0]
	}
	return queryParams, nil
}
