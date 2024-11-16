package middleware

import (
	"go.uber.org/zap"
)

// Middleware is a collection of general purpose request interceptors for the api servers
type Middleware struct {
	log *zap.SugaredLogger
}

// NewMiddleware constructs a new Middleware
func NewMiddleware(log *zap.SugaredLogger) *Middleware {
	return &Middleware{log: log}
}
