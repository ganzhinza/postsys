package memdb

import (
	"context"
	"postsys/internal/entity"
	"postsys/internal/errors"
	"sort"
	"sync"
	"time"
)

func New() *DB {
	return &DB{
		posts:              make(map[int32]*entity.Post),
		comments:           make(map[int32]*entity.Comment),
		rootCommentsByPost: make(map[int32][]*entity.Comment),
		childComments:      make(map[int32][]*entity.Comment),
		nextID: struct {
			post    int32
			comment int32
		}{post: 1, comment: 1},
	}
}

type DB struct {
	mu sync.RWMutex

	posts    map[int32]*entity.Post
	comments map[int32]*entity.Comment

	rootCommentsByPost map[int32][]*entity.Comment
	childComments      map[int32][]*entity.Comment

	nextID struct {
		post    int32
		comment int32
	}
}

func (db *DB) CreatePost(ctx context.Context, post entity.InputPost) (entity.Post, error) {
	fullPost := entity.Post{
		AuthorID:      post.AuthorID,
		Title:         post.Title,
		Content:       post.Content,
		AllowComments: post.AllowComments,
		CreatedAt:     time.Now(),
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	fullPost.ID = db.nextID.post
	db.nextID.post++

	db.posts[fullPost.ID] = &fullPost

	return fullPost, nil
}

func (db *DB) CreateComment(ctx context.Context, comment entity.InputComment, path []int32) (entity.Comment, error) {
	fullComment := entity.Comment{
		Content:   comment.Content,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		AuthorID:  comment.AuthorID,
		CreatedAt: time.Now(),
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	if comment.ParentID == nil {
		db.rootCommentsByPost[comment.PostID] = append(db.rootCommentsByPost[comment.PostID], &fullComment)
	} else {
		db.childComments[*comment.ParentID] = append(db.childComments[*comment.ParentID], &fullComment)
	}

	fullComment.ID = db.nextID.comment
	db.nextID.comment++

	fullComment.Path = append(path, fullComment.ID)
	db.comments[fullComment.ID] = &fullComment

	return fullComment, nil
}

func (db *DB) GetPosts(ctx context.Context) ([]entity.Post, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	posts := make([]entity.Post, 0, len(db.posts))
	for _, p := range db.posts {
		posts = append(posts, *p)
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].CreatedAt != posts[j].CreatedAt {
			return posts[i].CreatedAt.After(posts[j].CreatedAt)
		} else {
			return posts[i].ID > posts[j].ID
		}
	})

	return posts, nil
}

func (db *DB) GetPost(ctx context.Context, postID int32) (entity.Post, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	p, ok := db.posts[postID]
	if !ok {
		return entity.Post{}, errors.PostNotFoundError
	}
	return *p, nil
}

func (db *DB) GetRootComments(ctx context.Context, postID int32, limit, offset int32) ([]entity.Comment, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	roots := db.rootCommentsByPost[postID]
	if len(roots) == 0 {
		return []entity.Comment{}, nil
	}

	sorted := make([]*entity.Comment, len(roots))
	copy(sorted, roots)
	sort.Slice(sorted, func(i, j int) bool {
		return compareComments(sorted[i], sorted[j]) < 0
	})

	comments := make([]entity.Comment, 0)
	for i := offset; i < offset+limit && i < int32(len(sorted)); i++ {
		comments = append(comments, *sorted[i])
	}

	return comments, nil
}

func (db *DB) CountRootComments(ctx context.Context, postID int32) (int32, error) {
	return int32(len(db.rootCommentsByPost[postID])), nil
}

func (db *DB) GetCommentAvailability(ctx context.Context, postID int32) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	post, ok := db.posts[postID]
	if !ok {
		return false, errors.PostNotFoundError
	}

	return post.AllowComments, nil
}

func (db *DB) GetCommentPath(ctx context.Context, commentID int32) ([]int32, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	originalComment, ok := db.comments[commentID]
	if !ok {
		return nil, errors.CommentNotFoundError
	}

	path := make([]int32, len(originalComment.Path))
	copy(path, originalComment.Path)

	return path, nil
}

func (db *DB) UpdateCommentAvailability(ctx context.Context, postID int32, availability bool) (entity.Post, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	p, ok := db.posts[postID]
	if !ok {
		return entity.Post{}, errors.PostNotFoundError
	}
	p.AllowComments = availability

	return *p, nil
}

func (db *DB) GetCommentsByRootIDs(ctx context.Context, postID int32, rootIDs []int32) ([]entity.Comment, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var comments []entity.Comment

	for _, rootID := range rootIDs {
		db.collectSubtree(rootID, &comments)
	}

	sort.Slice(comments, func(i, j int) bool {
		return compareComments(&comments[i], &comments[j]) < 0
	})

	return comments, nil
}

func (db *DB) collectSubtree(parentID int32, comments *[]entity.Comment) {
	if c, ok := db.comments[parentID]; ok {
		*comments = append(*comments, *c)
	}
	for _, child := range db.childComments[parentID] {
		db.collectSubtree(child.ID, comments)
	}
}

func compareComments(a, b *entity.Comment) int {
	minLen := len(a.Path)
	if len(b.Path) < minLen {
		minLen = len(b.Path)
	}
	for i := 0; i < minLen; i++ {
		if a.Path[i] != b.Path[i] {
			if a.Path[i] < b.Path[i] {
				return -1
			}
			return 1
		}
	}

	if len(a.Path) != len(b.Path) {
		if len(a.Path) < len(b.Path) {
			return -1
		}
		return 1
	}

	if a.CreatedAt.After(b.CreatedAt) {
		return -1
	}
	if a.CreatedAt.Before(b.CreatedAt) {
		return 1
	}
	return 0
}
