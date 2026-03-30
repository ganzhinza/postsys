package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"postsys/internal/db/pgsql/sqlc"
	"postsys/internal/entity"
	apperr "postsys/internal/errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type db struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func New(pool *pgxpool.Pool) *db {
	return &db{
		pool: pool,
		q:    sqlc.New(pool),
	}
}

func (s *db) WithTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *db) GetPosts(ctx context.Context) ([]entity.Post, error) {
	dbPosts, err := s.q.GetPosts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]entity.Post, len(dbPosts))
	for i, p := range dbPosts {
		result[i] = entity.Post{
			ID:            p.ID,
			AuthorID:      p.AuthorID,
			Title:         p.Title,
			Content:       p.Content,
			AllowComments: p.AllowComments.Bool,
			CreatedAt:     p.CreatedAt.Time,
		}
	}
	return result, nil
}

func (s *db) GetPost(ctx context.Context, postID int32) (entity.Post, error) {
	dbPost, err := s.q.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Post{}, apperr.PostNotFoundError
		}
		return entity.Post{}, err
	}
	return entity.Post{
		ID:            dbPost.ID,
		AuthorID:      dbPost.AuthorID,
		Title:         dbPost.Title,
		Content:       dbPost.Content,
		AllowComments: dbPost.AllowComments.Bool,
		CreatedAt:     dbPost.CreatedAt.Time,
	}, nil
}

func (s *db) CountRootComments(ctx context.Context, postID int32) (int32, error) {
	count, err := s.q.CountRootComments(ctx, postID)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (s *db) GetRootComments(ctx context.Context, postID int32, limit, offset int32) ([]entity.Comment, error) {
	rows, err := s.q.GetRootComments(ctx, sqlc.GetRootCommentsParams{
		PostID: postID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	comments := make([]entity.Comment, len(rows))
	for i, c := range rows {
		var parentID *int32
		if c.ParentID.Valid {
			pid := c.ParentID.Int32
			parentID = &pid
		}
		comments[i] = entity.Comment{
			ID:        c.ID,
			Content:   c.Content,
			PostID:    c.PostID,
			ParentID:  parentID,
			AuthorID:  c.AuthorID,
			CreatedAt: c.CreatedAt.Time,
			Path:      c.Path,
		}
	}
	return comments, nil
}

func (s *db) GetCommentsByRootIDs(ctx context.Context, postID int32, rootIDs []int32) ([]entity.Comment, error) {
	rows, err := s.q.GetBranches(ctx, sqlc.GetBranchesParams{
		PostID:  postID,
		RootIds: rootIDs,
	})
	if err != nil {
		return nil, err
	}
	comments := make([]entity.Comment, len(rows))
	for i, c := range rows {
		var parentID *int32
		if c.ParentID.Valid {
			pid := c.ParentID.Int32
			parentID = &pid
		}
		comments[i] = entity.Comment{
			ID:        c.ID,
			Content:   c.Content,
			PostID:    c.PostID,
			ParentID:  parentID,
			AuthorID:  c.AuthorID,
			CreatedAt: c.CreatedAt.Time,
			Path:      c.Path,
		}
	}
	return comments, nil
}

func (s *db) GetCommentAvailability(ctx context.Context, postID int32) (bool, error) {
	allow, err := s.q.GetCommentsAvailability(ctx, int32(postID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, apperr.PostNotFoundError
		}
		return false, err
	}
	return allow.Bool, nil
}

func (s *db) GetCommentPath(ctx context.Context, commentID int32) ([]int32, error) {
	dbPath, err := s.q.GetCommentPath(ctx, int32(commentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.PathNotFoundError
		}
		return nil, err
	}

	return dbPath, nil
}

func (s *db) UpdateCommentAvailability(ctx context.Context, postID int32, availability bool) (entity.Post, error) {
	dbPost, err := s.q.UpdateCommentAvailability(ctx, sqlc.UpdateCommentAvailabilityParams{
		ID:            postID,
		AllowComments: pgtype.Bool{Bool: availability, Valid: true},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Post{}, apperr.PostNotFoundError
		}
		return entity.Post{}, err
	}
	return entity.Post{
		ID:            dbPost.ID,
		AuthorID:      dbPost.AuthorID,
		Title:         dbPost.Title,
		Content:       dbPost.Content,
		AllowComments: dbPost.AllowComments.Bool,
		CreatedAt:     dbPost.CreatedAt.Time,
	}, nil
}

func (s *db) CreatePost(ctx context.Context, post entity.InputPost) (entity.Post, error) {
	dbPost, err := s.q.CreatePost(ctx, sqlc.CreatePostParams{
		AuthorID:      post.AuthorID,
		Title:         post.Title,
		Content:       post.Content,
		AllowComments: pgtype.Bool{Bool: post.AllowComments, Valid: true},
	})
	if err != nil {
		return entity.Post{}, err
	}

	return entity.Post{
		ID:            dbPost.ID,
		AuthorID:      dbPost.AuthorID,
		Title:         dbPost.Title,
		Content:       dbPost.Content,
		AllowComments: dbPost.AllowComments.Bool,
		CreatedAt:     dbPost.CreatedAt.Time,
	}, nil
}

func (s *db) CreateComment(ctx context.Context, comment entity.InputComment, path []int32) (entity.Comment, error) {
	var result entity.Comment

	err := s.WithTx(ctx, func(q *sqlc.Queries) error {
		var parentID pgtype.Int4
		if comment.ParentID != nil {
			parentID = pgtype.Int4{Int32: *comment.ParentID, Valid: true}
		}

		dbComment, err := q.CreateComment(ctx, sqlc.CreateCommentParams{
			Content:  comment.Content,
			PostID:   comment.PostID,
			ParentID: parentID,
			AuthorID: comment.AuthorID,
			Path:     path,
		})
		if err != nil {
			return err
		}

		correctPath := append(dbComment.Path, dbComment.ID)
		updatedComment, err := q.UpdateCommentPath(ctx, sqlc.UpdateCommentPathParams{
			ID:   dbComment.ID,
			Path: correctPath,
		})
		if err != nil {
			return err
		}

		var parentIDRes *int32
		if updatedComment.ParentID.Valid {
			pid := updatedComment.ParentID.Int32
			parentIDRes = &pid
		}

		result = entity.Comment{
			ID:        updatedComment.ID,
			Content:   updatedComment.Content,
			PostID:    updatedComment.PostID,
			ParentID:  parentIDRes,
			AuthorID:  updatedComment.AuthorID,
			CreatedAt: updatedComment.CreatedAt.Time,
			Path:      updatedComment.Path,
		}
		return nil
	})

	return result, err
}
