package httputils

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
func GetQueryParams(query *url.Values) (map[string]interface{}, error) {
	queryParams := map[string]interface{}{}
	for key, value := range *query {
		if len(value) > 1 {
			err := QueryError{Filed: key, Detail: MultipleQueryValues}
			return nil, err
		}
		queryParams[key] = value[0]
	}
	if queryParams["limit"] == nil {
		queryParams["limit"] = DefaultPaginationLimitStr
	}
	limit, err := strconv.ParseUint(queryParams["limit"].(string), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid limit value. expected unsigned integer got %s", queryParams["limit"])
	}
	queryParams["limit"] = limit

	if queryParams["offset"] == nil {
		queryParams["offset"] = DefaultPaginationOffsetStr
	}
	offset, err := strconv.ParseUint(queryParams["offset"].(string), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid offset value. expected unsigned integer got %s", queryParams["offset"])
	}
	queryParams["offset"] = offset

	if queryParams["sort"] != nil {
		sortFields := queryParams["sort"]
		sortFieldsParts := strings.Split(sortFields.(string), ",")
		for _, sortField := range sortFieldsParts {
			sortFieldParts := strings.Split(sortField, ":")
			if len(sortFieldParts) == 2 {
				if sortFieldParts[1] != "asc" && sortFieldParts[1] != "desc" {
					return nil, fmt.Errorf("invalid sort order. expected asc or desc, got %s", sortFieldParts[1])
				}
			} else {
				return nil, fmt.Errorf("invalid sort value. expected field:how, got %s", sortField)
			}
		}

		queryParams["sort"] = sortFields
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
