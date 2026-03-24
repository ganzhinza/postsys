package service_test

import (
	"context"
	"postsys/internal/db/memdb"
	"postsys/internal/entity"
	"postsys/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupServiceTest(t *testing.T) (*service.ServiceImpl, *memdb.DB, int32, int32, int32, int32, int32, int32) {
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

	post3, err := db.CreatePost(ctx, entity.InputPost{
		AuthorID:      3,
		Title:         "Post 3",
		Content:       "Content 3",
		AllowComments: false,
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

	svc := service.New(db)
	return svc, db, post1.ID, post2.ID, post3.ID, root1.ID, root2.ID, child.ID
}

func TestService_GetPosts(t *testing.T) {
	svc, _, _, _, _, _, _, _ := setupServiceTest(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		wantCount   int
		expectedIDs []int32
	}{
		{
			name:        "all posts sorted",
			wantCount:   3,
			expectedIDs: []int32{3, 2, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := svc.GetPosts(ctx)
			require.NoError(t, err)
			assert.Len(t, posts, tt.wantCount)
			for i, wantID := range tt.expectedIDs {
				assert.Equal(t, wantID, posts[i].ID)
			}
		})
	}
}

func TestService_GetPost(t *testing.T) {
	svc, _, post1ID, _, _, _, _, _ := setupServiceTest(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		postID  int32
		wantErr bool
	}{
		{"existing post", post1ID, false},
		{"not found", 999, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetPost(ctx, tt.postID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.postID, got.ID)
		})
	}
}

func TestService_GetCommentsTree(t *testing.T) {
	svc, _, post1ID, _, _, root1ID, root2ID, childID := setupServiceTest(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		limit    int32
		offset   int32
		wantTree *entity.CommentTree
	}{
		{
			name:   "first page, limit 1",
			limit:  1,
			offset: 0,
			wantTree: &entity.CommentTree{
				Roots: []entity.Comment{
					{ID: root1ID, Content: "Root comment 1", PostID: post1ID, AuthorID: 1},
				},
				Children: map[int32][]entity.Comment{
					root1ID: {
						{ID: childID, Content: "Child comment", PostID: post1ID, AuthorID: 1, ParentID: &root1ID},
					},
				},
				TotalRoots: 2,
			},
		},
		{
			name:   "second page, limit 1",
			limit:  1,
			offset: 1,
			wantTree: &entity.CommentTree{
				Roots: []entity.Comment{
					{ID: root2ID, Content: "Root comment 2", PostID: post1ID, AuthorID: 1},
				},
				Children:   map[int32][]entity.Comment{},
				TotalRoots: 2,
			},
		},
		{
			name:   "offset beyond last",
			limit:  10,
			offset: 10,
			wantTree: &entity.CommentTree{
				Roots:      []entity.Comment{},
				Children:   map[int32][]entity.Comment{},
				TotalRoots: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := svc.GetCommentsTree(ctx, post1ID, tt.limit, tt.offset)
			require.NoError(t, err)
			assertCommentTreeEqual(t, tt.wantTree, tree)
		})
	}
}

func TestService_CreatePost(t *testing.T) {
	tests := []struct {
		name    string
		input   entity.InputPost
		wantID  int32
		wantErr bool
	}{
		{
			name: "valid post",
			input: entity.InputPost{
				AuthorID:      4,
				Title:         "New Post",
				Content:       "New Content",
				AllowComments: true,
			},
			wantID: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _, _, _, _, _, _ := setupServiceTest(t)
			ctx := context.Background()

			got, err := svc.CreatePost(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
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

func TestService_CreateComment(t *testing.T) {
	svc, _, post1ID, _, postDisabledID, root1ID, _, _ := setupServiceTest(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		input        entity.InputComment
		wantErr      bool
		errMsg       string
		expectedPath []int32
	}{
		{
			name: "valid root comment",
			input: entity.InputComment{
				PostID:   post1ID,
				Content:  "New root",
				ParentID: nil,
				AuthorID: 2,
			},
			expectedPath: []int32{4},
		},
		{
			name: "valid child comment",
			input: entity.InputComment{
				PostID:   post1ID,
				Content:  "New child",
				ParentID: &root1ID,
				AuthorID: 2,
			},
			expectedPath: []int32{1, 5},
		},
		{
			name: "comment too long",
			input: entity.InputComment{
				PostID:   post1ID,
				Content:  string(make([]byte, 2001)),
				ParentID: nil,
				AuthorID: 2,
			},
			wantErr: true,
			errMsg:  "comment too long",
		},
		{
			name: "comments disabled",
			input: entity.InputComment{
				PostID:   postDisabledID,
				Content:  "Should fail",
				ParentID: nil,
				AuthorID: 2,
			},
			wantErr: true,
			errMsg:  "comments disabled",
		},
		{
			name: "parent not found",
			input: entity.InputComment{
				PostID:   post1ID,
				Content:  "Invalid parent",
				ParentID: intPtr(999),
				AuthorID: 2,
			},
			wantErr: true,
			errMsg:  "comment not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CreateComment(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input.Content, got.Content)
			assert.Equal(t, tt.input.PostID, got.PostID)
			if tt.input.ParentID == nil {
				assert.Nil(t, got.ParentID)
			} else {
				assert.NotNil(t, got.ParentID)
				assert.Equal(t, *tt.input.ParentID, *got.ParentID)
			}
			if tt.expectedPath != nil {
				assert.Equal(t, tt.expectedPath, got.Path)
			}
		})
	}
}

func TestService_UpdateCommentAvailability(t *testing.T) {
	svc, db, post1ID, _, postDisabledID, _, _, _ := setupServiceTest(t)
	ctx := context.Background()

	tests := []struct {
		name            string
		postID          int32
		userID          int32
		newAvailability bool
		wantErr         bool
		errMsg          string
		expectedState   bool
	}{
		{
			name:            "owner disables comments",
			postID:          post1ID,
			userID:          1,
			newAvailability: false,
			expectedState:   false,
		},
		{
			name:            "owner enables comments",
			postID:          postDisabledID,
			userID:          3,
			newAvailability: true,
			expectedState:   true,
		},
		{
			name:            "non-owner cannot update",
			postID:          post1ID,
			userID:          2,
			newAvailability: false,
			wantErr:         true,
			errMsg:          "only post owner",
		},
		{
			name:            "post not found",
			postID:          999,
			userID:          1,
			newAvailability: true,
			wantErr:         true,
			errMsg:          "post not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := svc.UpdateCommentAvailability(ctx, tt.postID, tt.userID, tt.newAvailability)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.postID, updated.ID)
			assert.Equal(t, tt.expectedState, updated.AllowComments)

			avail, err := db.GetCommentAvailability(ctx, tt.postID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedState, avail)
		})
	}
}

func intPtr(i int32) *int32 {
	return &i
}

func assertCommentTreeEqual(t *testing.T, want, got *entity.CommentTree) {
	t.Helper()

	require.Equal(t, len(want.Roots), len(got.Roots), "roots length mismatch")
	for i := range want.Roots {
		assertCommentEqual(t, want.Roots[i], got.Roots[i])
	}

	require.Equal(t, len(want.Children), len(got.Children), "children map size mismatch")
	for parentID, wantChildren := range want.Children {
		gotChildren, ok := got.Children[parentID]
		require.True(t, ok, "missing children for parent %d", parentID)
		require.Equal(t, len(wantChildren), len(gotChildren), "children count mismatch for parent %d", parentID)
		for j := range wantChildren {
			assertCommentEqual(t, wantChildren[j], gotChildren[j])
		}
	}
	assert.Equal(t, want.TotalRoots, got.TotalRoots)
}

func assertCommentEqual(t *testing.T, want, got entity.Comment) {
	t.Helper()
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Content, got.Content)
	assert.Equal(t, want.PostID, got.PostID)
	assert.Equal(t, want.AuthorID, got.AuthorID)
	if want.ParentID == nil {
		assert.Nil(t, got.ParentID)
	} else {
		assert.NotNil(t, got.ParentID)
		assert.Equal(t, *want.ParentID, *got.ParentID)
	}
}
