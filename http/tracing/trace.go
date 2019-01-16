package tracing

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/microdevs/missy/log"
)

const (
	// RequestIDHeader defines key for the request id.
	RequestIDHeader string = "X-Request-Id"

	loggerRequestIDKey string = "requestId"
	loggerIPKey        string = "ip"
	handlerKey         string = "handler"
)

// RequestID allows to get request id from context.
func RequestID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

// FromRequest wraps logger with fields: request id, ip, handlers.
func FromRequest(l log.FieldsLogger, handler []string, r *http.Request) log.FieldsLogger {
	l = l.WithField(loggerIPKey, r.RemoteAddr)
	l = l.WithField(handlerKey, handler)
	l = FromContext(l, r.Context())
	return l
}

// FromContext adds request id to the logger fields.
func FromContext(l log.FieldsLogger, ctx context.Context) log.FieldsLogger {
	reqID := RequestID(ctx)
	if reqID != "" {
		return l.WithField(loggerRequestIDKey, reqID)
	}
	return l
}

// HTTPReqWithID adds request id to the request headers.
func HTTPReqWithID(ctx context.Context, r *http.Request) {
	reqID := RequestID(ctx)
	if reqID != "" {
		r.Header.Set(RequestIDHeader, reqID)
	}
}
