package adapter

import (
	"postsys/internal/entity"
	"postsys/internal/graph/model"
)

func ToServiceInputPost(in model.InputPost) entity.InputPost {
	return entity.InputPost{
		AuthorID:      in.AuthorID,
		Title:         in.Title,
		Content:       in.Content,
		AllowComments: in.AllowComments,
	}
}

func ToServiceInputComment(in model.InputComment) entity.InputComment {
	return entity.InputComment{
		Content:  in.Content,
		AuthorID: in.AuthorID,
		PostID:   in.PostID,
		ParentID: in.ParentID,
	}
}

func ToGraphQLComment(c entity.Comment) model.Comment {
	return model.Comment{
		ID:        c.ID,
		Content:   c.Content,
		PostID:    c.PostID,
		ParentID:  c.ParentID,
		AuthorID:  c.AuthorID,
		CreatedAt: c.CreatedAt,
	}
}

func ToGraphQLPost(p entity.Post) model.Post {
	return model.Post{
		ID:            p.ID,
		AuthorID:      p.AuthorID,
		Title:         p.Title,
		Content:       p.Content,
		AllowComments: p.AllowComments,
		CreatedAt:     p.CreatedAt,
	}
}
