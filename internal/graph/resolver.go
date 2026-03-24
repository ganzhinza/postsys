package graph

import (
	"postsys/internal/graph/model"
	"postsys/internal/service"
	"sync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	service            service.Service
	subscribersByPosts map[int32][]chan<- *model.Comment
	mu                 sync.RWMutex
}

func NewResolver(srv service.Service) *Resolver {
	return &Resolver{
		service:            srv,
		subscribersByPosts: make(map[int32][]chan<- *model.Comment),
	}
}
