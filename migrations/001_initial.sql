-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    allow_comments BOOLEAN,
    created_at TIMESTAMP  WITH TIME ZONE NOT NULL DEFAULT NOW(),
    author_id INT NOT NULL
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    content VARCHAR(2000) NOT NULL,
    post_id INT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    parent_id INT REFERENCES comments(id) ON DELETE CASCADE,
    author_id INT NOT NULL,
    created_at TIMESTAMP  WITH TIME ZONE NOT NULL DEFAULT NOW(),
    path  INT[] NOT NULL
);

CREATE INDEX idx_comments_post_id ON comments USING HASH (post_id);
CREATE INDEX idx_comments_parent_id ON comments USING HASH (parent_id);
CREATE INDEX idx_comments_post_path_time ON comments (post_id, path, created_at DESC);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE posts;
DROP TABLE comments;
-- +goose StatementEnd