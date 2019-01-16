package middleware

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

// Tracing allows to set request ID for any incoming HTTP request.
// RequestID will be set at request's context.
func Tracing(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}
