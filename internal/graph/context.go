package graph

import (
	"context"
	"net/http"
	"postsys/internal/entity"
)

type contextKey string

const childrenMapKey contextKey = "childrenMap"

func WithChildrenMap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), childrenMapKey, make(map[int32][]entity.Comment))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
