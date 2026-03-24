package memdb_test

import (
	"context"
	"postsys/internal/db/memdb"
	"postsys/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test db
func setupDB(t *testing.T) (*memdb.DB, int32, int32, int32, int32, int32) {
	db := memdb.New()
	ctx := context.Background()

	post1, err := db.CreatePost(ctx, entity.InputPost{
		AuthorID:      1,
		Title:         "Post 1",
		Content:       "Content 1",
		AllowComments: true,
	})
	require.NoError(t, err)

	post2, err := db.CreatePost(ctx, entity.InputPost{
		AuthorID:      2,
		Title:         "Post 2",
		Content:       "Content 2",
		AllowComments: true,
	})
	require.NoError(t, err)

	root1, err := db.CreateComment(ctx, entity.InputComment{
		Content:  "Root comment 1",
		PostID:   post1.ID,
		ParentID: nil,
		AuthorID: 1,
	}, []int32{})
	require.NoError(t, err)

	root2, err := db.CreateComment(ctx, entity.InputComment{
		Content:  "Root comment 2",
		PostID:   post1.ID,
		ParentID: nil,
		AuthorID: 1,
	}, []int32{})
	require.NoError(t, err)

	child, err := db.CreateComment(ctx, entity.InputComment{
		Content:  "Child comment",
		PostID:   post1.ID,
		ParentID: &root1.ID,
		AuthorID: 1,
	}, root1.Path)
	require.NoError(t, err)

	return db, post1.ID, post2.ID, root1.ID, root2.ID, child.ID
}

func TestDB_CreatePost(t *testing.T) {
	tests := []struct {
		name    string
		input   entity.InputPost
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid post",
			input: entity.InputPost{
				AuthorID:      1,
				Title:         "Test Post",
				Content:       "Test Content",
				AllowComments: true,
			},
			wantID:  1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := memdb.New()
			ctx := context.Background()

			got, err := db.CreatePost(ctx, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got.ID)
			assert.Equal(t, tt.input.Title, got.Title)
			assert.Equal(t, tt.input.Content, got.Content)
			assert.Equal(t, tt.input.AuthorID, got.AuthorID)
			assert.Equal(t, tt.input.AllowComments, got.AllowComments)
			assert.NotZero(t, got.CreatedAt)
		})
	}
}

func TestDB_CreateComment(t *testing.T) {
	db, postID1, _, root1ID, _, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		input      entity.InputComment
		path       []int32
		wantID     int32
		wantParent *int32
		wantPath   []int32
	}{
		{
			name: "new root comment",
			input: entity.InputComment{
				Content:  "New root",
				PostID:   postID1,
				ParentID: nil,
				AuthorID: 2,
			},
			path:       []int32{},
			wantID:     4,
			wantParent: nil,
			wantPath:   []int32{4},
		},
		{
			name: "new child of root1",
			input: entity.InputComment{
				Content:  "New child",
				PostID:   postID1,
				ParentID: &root1ID,
				AuthorID: 2,
			},
			path:       []int32{root1ID},
			wantID:     5,
			wantParent: &root1ID,
			wantPath:   []int32{root1ID, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.CreateComment(ctx, tt.input, tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got.ID)
			assert.Equal(t, tt.input.Content, got.Content)
			assert.Equal(t, tt.input.PostID, got.PostID)
			assert.Equal(t, tt.input.AuthorID, got.AuthorID)
			if tt.wantParent == nil {
				assert.Nil(t, got.ParentID)
			} else {
				assert.NotNil(t, got.ParentID)
				assert.Equal(t, *tt.wantParent, *got.ParentID)
			}
			assert.Equal(t, tt.wantPath, got.Path)
			assert.NotZero(t, got.CreatedAt)
		})
	}
}
func TestDB_GetPosts(t *testing.T) {
	db, _, _, _, _, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		expectedIDs []int32 // from new to old
	}{
		{
			name:        "posts sorted",
			expectedIDs: []int32{2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := db.GetPosts(ctx)
			require.NoError(t, err)
			require.Len(t, posts, len(tt.expectedIDs))

			for i, wantID := range tt.expectedIDs {
				assert.Equal(t, wantID, posts[i].ID)
			}
		})
	}
}

func TestDB_GetPost(t *testing.T) {
	db, postID1, _, _, _, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		postID    int32
		wantTitle string
		wantErr   bool
	}{
		{"existing post", postID1, "Post 1", false},
		{"not found", 999, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetPost(ctx, tt.postID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.postID, got.ID)
			assert.Equal(t, tt.wantTitle, got.Title)
		})
	}
}

func TestDB_GetRootComments(t *testing.T) {
	db, postID1, _, root1, root2, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		postID      int32
		limit       int32
		offset      int32
		expectedLen int
		expectedIDs []int32
	}{
		{
			name:        "all roots",
			postID:      postID1,
			limit:       10,
			offset:      0,
			expectedLen: 2,
			expectedIDs: []int32{root1, root2},
		},
		{
			name:        "first page with limit 1",
			postID:      postID1,
			limit:       1,
			offset:      0,
			expectedLen: 1,
			expectedIDs: []int32{root1},
		},
		{
			name:        "second page with limit 1",
			postID:      postID1,
			limit:       1,
			offset:      1,
			expectedLen: 1,
			expectedIDs: []int32{root2},
		},
		{
			name:        "offset beyond last",
			postID:      postID1,
			limit:       1,
			offset:      2,
			expectedLen: 0,
			expectedIDs: []int32{},
		},
		{
			name:        "post without comments",
			postID:      999,
			limit:       10,
			offset:      0,
			expectedLen: 0,
			expectedIDs: []int32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots, err := db.GetRootComments(ctx, tt.postID, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Len(t, roots, tt.expectedLen)

			ids := make([]int32, len(roots))
			for i, r := range roots {
				ids[i] = r.ID
			}
			assert.Equal(t, tt.expectedIDs, ids)
		})
	}
}

func TestDB_CountRootComments(t *testing.T) {
	db, postID1, postID2, _, _, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		postID int32
		want   int32
	}{
		{
			name:   "post with comments",
			postID: postID1,
			want:   2,
		},
		{
			name:   "post without comments",
			postID: postID2,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := db.CountRootComments(ctx, tt.postID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, count)
		})
	}
}

func TestDB_GetCommentAvailability(t *testing.T) {
	db, postID1, _, _, _, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		postID  int32
		wantErr bool
		want    bool
		errMsg  string
	}{
		{
			name:    "existing post",
			postID:  postID1,
			wantErr: false,
			want:    true,
		},
		{
			name:    "non-existent post",
			postID:  999,
			wantErr: true,
			errMsg:  "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetCommentAvailability(ctx, tt.postID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDB_GetCommentPath(t *testing.T) {
	db, _, _, rootID, _, childID := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		commentID int32
		wantPath  []int32
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "root comment path",
			commentID: rootID,
			wantPath:  []int32{rootID},
			wantErr:   false,
		},
		{
			name:      "child comment path",
			commentID: childID,
			wantPath:  []int32{rootID, childID},
			wantErr:   false,
		},
		{
			name:      "not found",
			commentID: 999,
			wantErr:   true,
			errMsg:    "comment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := db.GetCommentPath(ctx, tt.commentID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, path)
		})
	}
}

func TestDB_UpdateCommentAvailability(t *testing.T) {
	db, postID1, postID2, _, _, _ := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name            string
		postID          int32
		newAvailability bool
		wantErr         bool
		errMsg          string
		expectedState   bool
	}{
		{
			name:            "disable comments on existing post",
			postID:          postID1,
			newAvailability: false,
			wantErr:         false,
			expectedState:   false,
		},
		{
			name:            "enable comments on existing post",
			postID:          postID2,
			newAvailability: true,
			wantErr:         false,
			expectedState:   true,
		},
		{
			name:            "non-existent post",
			postID:          999,
			newAvailability: true,
			wantErr:         true,
			errMsg:          "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedPost, err := db.UpdateCommentAvailability(ctx, tt.postID, tt.newAvailability)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.postID, updatedPost.ID)
			assert.Equal(t, tt.expectedState, updatedPost.AllowComments)

			available, err := db.GetCommentAvailability(ctx, tt.postID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedState, available)
		})
	}
}

func TestDB_GetCommentsByRootIDs(t *testing.T) {
	db, postID1, _, root1, root2, child := setupDB(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		rootIDs     []int32
		expectedLen int
		expectedIDs []int32
	}{
		{
			name:        "single root with children",
			rootIDs:     []int32{root1},
			expectedLen: 2,
			expectedIDs: []int32{root1, child},
		},
		{
			name:        "single root without children",
			rootIDs:     []int32{root2},
			expectedLen: 1,
			expectedIDs: []int32{root2},
		},
		{
			name:        "multiple roots",
			rootIDs:     []int32{root1, root2},
			expectedLen: 3,
			expectedIDs: []int32{root1, child, root2},
		},
		{
			name:        "empty root list",
			rootIDs:     []int32{},
			expectedLen: 0,
			expectedIDs: []int32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := db.GetCommentsByRootIDs(ctx, postID1, tt.rootIDs)
			require.NoError(t, err)
			assert.Len(t, comments, tt.expectedLen)

			ids := make([]int32, len(comments))
			for i, c := range comments {
				ids[i] = c.ID
			}
			assert.Equal(t, tt.expectedIDs, ids)
		})
	}
}
