package db

import (
	"context"
	"postsys/internal/entity"
)

type DB interface {
	GetPosts(ctx context.Context) ([]entity.Post, error)
	GetPost(ctx context.Context, postID int32) (entity.Post, error)

	GetRootComments(ctx context.Context, postID int32, limit, offset int32) ([]entity.Comment, error)
	CountRootComments(ctx context.Context, postID int32) (int32, error)

	GetCommentsByRootIDs(ctx context.Context, postID int32, rootIDs []int32) ([]entity.Comment, error)

	GetCommentAvailability(ctx context.Context, postID int32) (bool, error)
	GetCommentPath(ctx context.Context, commentID int32) ([]int32, error)

	UpdateCommentAvailability(ctx context.Context, postID int32, availability bool) (entity.Post, error)

	CreatePost(ctx context.Context, post entity.InputPost) (entity.Post, error)
	CreateComment(ctx context.Context, comment entity.InputComment, path []int32) (entity.Comment, error)
}
