package service

import (
	"context"
	"postsys/internal/db"
	"postsys/internal/entity"
	"postsys/internal/errors"
)

type ServiceImpl struct {
	db db.DB
}

func New(db db.DB) *ServiceImpl {
	return &ServiceImpl{
		db: db,
	}
}

func (s *ServiceImpl) GetPosts(ctx context.Context) ([]entity.Post, error) {
	posts, err := s.db.GetPosts(ctx)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *ServiceImpl) GetPost(ctx context.Context, id int32) (entity.Post, error) {
	return s.db.GetPost(ctx, id)
}

func (s *ServiceImpl) GetCommentsTree(ctx context.Context, postID, limit, offset int32) (*entity.CommentTree, error) {
	roots, err := s.db.GetRootComments(ctx, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	totalRoots, err := s.db.CountRootComments(ctx, postID)
	if err != nil {
		return nil, err
	}

	if len(roots) == 0 {
		return &entity.CommentTree{
			Roots:      []entity.Comment{},
			Children:   map[int32][]entity.Comment{},
			TotalRoots: totalRoots,
		}, nil
	}

	rootIDs := make([]int32, len(roots))
	for i, r := range roots {
		rootIDs[i] = r.ID
	}

	allComments, err := s.db.GetCommentsByRootIDs(ctx, postID, rootIDs)
	if err != nil {
		return nil, err
	}

	childrenMap := buildCommentsMap(allComments)

	return &entity.CommentTree{
		Roots:      roots,
		Children:   childrenMap,
		TotalRoots: totalRoots,
	}, nil
}

func (s *ServiceImpl) CreatePost(ctx context.Context, post entity.InputPost) (entity.Post, error) {
	return s.db.CreatePost(ctx, post)
}

func (s *ServiceImpl) CreateComment(ctx context.Context, comment entity.InputComment) (entity.Comment, error) {
	availability, err := s.db.GetCommentAvailability(ctx, comment.PostID)
	if err != nil {
		return entity.Comment{}, err
	}
	if availability == false {
		return entity.Comment{}, errors.CommentsDisabledError
	}
	if len(comment.Content) > 2000 {
		return entity.Comment{}, errors.CommentTooLongError
	}

	path := []int32{}
	if comment.ParentID != nil {
		path, err = s.db.GetCommentPath(ctx, *comment.ParentID)
		if err != nil {
			return entity.Comment{}, err
		}
	}

	resComment, err := s.db.CreateComment(ctx, comment, path)
	if err != nil {
		return entity.Comment{}, err
	}
	return resComment, nil
}

func (s *ServiceImpl) UpdateCommentAvailability(ctx context.Context, postID, userID int32, availability bool) (entity.Post, error) {
	post, err := s.db.GetPost(ctx, postID)
	if err != nil {
		return entity.Post{}, err
	}
	if userID != post.AuthorID {
		return entity.Post{}, errors.NotOwnerUpdatesCommentAvailabilityError
	}
	return s.db.UpdateCommentAvailability(ctx, postID, availability)
}

func buildCommentsMap(comments []entity.Comment) map[int32][]entity.Comment {
	children := make(map[int32][]entity.Comment)
	for _, c := range comments {
		if c.ParentID != nil {
			pid := *c.ParentID
			children[pid] = append(children[pid], c)
		}
	}
	return children
}
