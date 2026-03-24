-- name: GetPosts :many
SELECT * FROM posts;

-- name: GetPost :one
SELECT * FROM posts WHERE id = $1;

-- name: GetRootComments :many
SELECT * FROM comments
WHERE post_id = $1 AND parent_id IS NULL
ORDER BY path, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountRootComments :one
SELECT COUNT(*) FROM comments
WHERE post_id = $1 AND parent_id IS NULL;

-- name: GetBranches :many
SELECT * FROM comments
WHERE post_id = $1
  AND path[1] = ANY(sqlc.arg(root_ids)::int[])
ORDER BY path, created_at DESC;

-- name: GetCommentsAvailability :one
SELECT allow_comments FROM posts WHERE id = $1;

-- name: GetCommentPath :one
SELECT path FROM comments WHERE id = $1;

-- name: UpdateCommentPath :one
UPDATE comments SET path = $2
WHERE id = $1 RETURNING *;

-- name: UpdateCommentAvailability :one
UPDATE posts SET allow_comments = $2 
WHERE id = $1 RETURNING *;

-- name: CreatePost :one
INSERT INTO posts (author_id, title, content, allow_comments) 
VALUES ($1, $2, $3, $4) 
RETURNING *;

-- name: CreateComment :one
INSERT INTO comments (content, post_id, parent_id, author_id, path) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;