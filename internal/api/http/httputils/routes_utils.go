package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/package/utils"
)

type APIResponse[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

type errorStruct struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIError APIResponse[errorStruct]

func ReturnError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	code := http.StatusText(statusCode)
	response := APIError{
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
	response := APIResponse[any]{
		Ok:   true,
		Data: data,
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
func GetQueryParams(query *url.Values) (map[string]any, error) {
	queryParams := map[string]any{}
	for key, value := range *query {
		if len(value) > 1 {
			return nil, QueryError{Filed: key, Detail: MultipleQueryValues}
		}
		queryParams[key] = value[0]
	}

	setDefault := func(param string, defaultValue string) {
		if queryParams[param] == nil {
			queryParams[param] = defaultValue
		}
	}

	setDefault("limit", DefaultPaginationLimitStr)
	setDefault("offset", DefaultPaginationOffsetStr)

	limit, err := strconv.ParseInt(queryParams["limit"].(string), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid limit value. expected unsigned integer got %s", queryParams["limit"])
	}
	queryParams["limit"] = int(limit)

	offset, err := strconv.ParseInt(queryParams["offset"].(string), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid offset value. expected unsigned integer got %s", queryParams["offset"])
	}
	queryParams["offset"] = int(offset)

	if sortFields, ok := queryParams["sort"]; ok {
		sortFieldsParts := strings.Split(sortFields.(string), ",")
		for _, sortField := range sortFieldsParts {
			sortFieldParts := strings.Split(sortField, ":")
			if len(sortFieldParts) != 2 || (sortFieldParts[1] != "asc" && sortFieldParts[1] != "desc") {
				return nil, fmt.Errorf("invalid sort value. expected field:how, got %s", sortField)
			}
		}
	} else {
		queryParams["sort"] = ""
	}

	return queryParams, nil
}

// SaveMultiPartFile saves an uploaded multipart file to a temporary directory and returns the file path.
func SaveMultiPartFile(file multipart.File, handler *multipart.FileHeader) (string, error) {
	tempDir := os.TempDir()

	filePath := fmt.Sprintf("%s/%s", tempDir, handler.Filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// ShouldBindJSON binds the request body to a struct and validates it.
func ShouldBindJSON(body io.ReadCloser, v any) error {
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&v)
	if err != nil {
		return err
	}

	if dec.More() {
		return errors.New("unexpected extra data in JSON body")
	}

	validator, err := utils.NewValidator()
	if err != nil {
		return err
	}
	if err := validator.Struct(v); err != nil {
		return err
	}
	return nil
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	ReturnError(w, http.StatusNotFound, "Requested route not found. Verify the URL is correct.")
}
