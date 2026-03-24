package entity

import "time"

type InputPost struct {
	AuthorID      int32
	Title         string
	Content       string
	AllowComments bool
}

type Post struct {
	ID            int32
	AuthorID      int32
	Title         string
	Content       string
	AllowComments bool
	CreatedAt     time.Time
}

type InputComment struct {
	Content  string
	AuthorID int32
	PostID   int32
	ParentID *int32
}

type Comment struct {
	ID        int32
	Content   string
	PostID    int32
	ParentID  *int32
	AuthorID  int32
	CreatedAt time.Time
	Path      []int32
}

type CommentTree struct {
	Roots      []Comment
	Children   map[int32][]Comment
	TotalRoots int32
}
