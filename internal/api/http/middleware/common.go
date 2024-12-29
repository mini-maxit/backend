package middleware

type ContextKey string

const (
	// SessionKey is the key used to store the session in the context.
	SessionKey ContextKey = "session"
	// UserIDKey is the key used to store the user ID in the context.
	UserIDKey ContextKey = "userId"
	// DatabaseKey is the key used to store the database connection in the context.
	DatabaseKey ContextKey = "database"
)
