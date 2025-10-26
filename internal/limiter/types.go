package limiter

import (
	"context"
	"net/http"
)

type LimiterType string

const (
	LimiterTypeLeaky    LimiterType = "leaky"
	LimiterTypeToken    LimiterType = "token"
	LimiterTypeWeighted LimiterType = "token_weighted"
)

type Limiter interface {
	Limit(ctx context.Context, id string, parameters ...any) bool
	Stop()
	Middleware(http.HandlerFunc) http.HandlerFunc
}
