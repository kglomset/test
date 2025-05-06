package middleware

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

// Logging response severity levels
const (
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

// LoggingHandler handles logging middleware
type LoggingHandler struct {
	db *sql.DB
}

// NewLoggingHandler creates a new logging handler
func NewLoggingHandler(db *sql.DB) *LoggingHandler {
	return &LoggingHandler{db: db}
}

// statusResponseWriter is a wrapper around http.ResponseWriter that keeps track of the status code.
type statusResponseWriter struct {
	statusCode int
	http.ResponseWriter
}

// WriteHeader writes the status code to the custom response writer.
func (srw *statusResponseWriter) WriteHeader(code int) {
	if srw.statusCode == 0 || srw.statusCode != code {
		srw.statusCode = code
		srw.ResponseWriter.WriteHeader(code)
	}
}

// Log logs the response logLevel, status code, status text and timestamp.
func (l *LoggingHandler) Log(level string, statusCode int) {
	// Logging the severity level of the users action, response status code, status text and timestamp.
	log.Printf("Response: [%s] %d %s, On: %s", level, statusCode, http.StatusText(statusCode),
		time.Now().Format(time.RFC1123))
}

// LoggingMiddleware is a middleware that makes it possible to
// perform logging of sever requests and calling the next handler.
func (l *LoggingHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logging the request method, path and timestamp.
		log.Printf("Request: %s %s, On: %s", r.Method, r.URL.Path, time.Now().Format(time.RFC1123))

		// Creating a new response writer for status codes and setting status code 200 as default.
		srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler.
		next.ServeHTTP(srw, r)

		// Logging the response severity level based on the status code.
		switch {
		case 500 <= srw.statusCode:
			l.Log(ERROR, srw.statusCode)
		case 400 <= srw.statusCode:
			l.Log(WARN, srw.statusCode)
		default:
			l.Log(INFO, srw.statusCode)
		}
	})
}
