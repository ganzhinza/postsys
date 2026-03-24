package service

import (
	"context"
	"postsys/internal/entity"
)

type Service interface {
	GetPosts(ctx context.Context) ([]entity.Post, error)
	GetPost(ctx context.Context, postID int32) (entity.Post, error)
	GetCommentsTree(ctx context.Context, postID int32, limit, offset int32) (*entity.CommentTree, error)

	CreatePost(ctx context.Context, post entity.InputPost) (entity.Post, error)
	CreateComment(ctx context.Context, comment entity.InputComment) (entity.Comment, error)

	UpdateCommentAvailability(ctx context.Context, postID, userID int32, availability bool) (entity.Post, error)
}
