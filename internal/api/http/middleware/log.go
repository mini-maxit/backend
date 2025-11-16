package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type wrappedResponseWriter struct {
	statusCode int
	http.ResponseWriter
}

func (lrw *wrappedResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs details of each HTTP request.
func LoggingMiddleware(next http.Handler, log *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedW := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedW, r)
		protocol := "http"
		if r.TLS != nil {
			protocol = "https"
		}
		log.Infof("%s %s %d host=%s service=%dms bytes=%d protocol=%s",
			r.Method,
			r.URL.Path,
			wrappedW.statusCode,
			r.URL.Hostname(),
			time.Since(start).Milliseconds(),
			r.ContentLength,
			protocol)
	})
}
