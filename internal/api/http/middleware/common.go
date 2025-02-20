package middleware

type ContextKey string

const (
	// SessionKey is the key used to store the session in the context.
	SessionKey ContextKey = "session"
	// UserIDKey is the key used to store the user ID in the context.
	UserIDKey ContextKey = "userId"
	// UserKey is the key used to store the user in the context.
	UserKey ContextKey = "user"
	// DatabaseKey is the key used to store the database connection in the context.
	DatabaseKey ContextKey = "database"
	// QueryParamsKey is the key used to store the query parameters of current request in the context.
	QueryParamsKey ContextKey = "queryParams"
)
