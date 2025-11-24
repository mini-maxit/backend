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
	"unicode"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/utils"
)

type APIResponse[T any] struct {
	Ok   bool `json:"ok"`
	Data T    `json:"data"`
}

func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(rune(s[i-1])) || unicode.IsDigit(rune(s[i-1]))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
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
		raw := sortFields.(string)
		if raw != "" {
			sortFieldsParts := strings.Split(raw, ",")
			converted := make([]string, 0, len(sortFieldsParts))
			for _, sortField := range sortFieldsParts {
				sortField = strings.TrimSpace(sortField)
				if sortField == "" {
					continue
				}
				sortFieldParts := strings.Split(sortField, ":")
				if len(sortFieldParts) != 2 || (sortFieldParts[1] != "asc" && sortFieldParts[1] != "desc") {
					return nil, fmt.Errorf("invalid sort value. expected field:how, got %s", sortField)
				}
				fieldName := sortFieldParts[0]
				fieldName = camelToSnake(fieldName)
				converted = append(converted, fieldName+":"+sortFieldParts[1])
			}
			queryParams["sort"] = strings.Join(converted, ",")
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
	ReturnError(w, http.StatusNotFound, "Endpoint not found")
}

// GetPathValue retrieves a path variable from the gorilla/mux request
func GetPathValue(r *http.Request, name string) string {
	if vars := mux.Vars(r); vars != nil {
		return vars[name]
	}
	return ""
}

func ExtractPaginationParams(queryParams map[string]any) schemas.PaginationParams {
	limit, ok := queryParams["limit"].(int)
	if !ok {
		limit = 20
	}
	offset, ok := queryParams["offset"].(int)
	if !ok {
		offset = 0
	}
	sort, ok := queryParams["sort"].(string)
	if !ok {
		sort = ""
	}
	return schemas.PaginationParams{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}
}

func GetCurrentUser(r *http.Request) *schemas.User {
	user := r.Context().Value(UserKey).(schemas.User)
	return &user
}
