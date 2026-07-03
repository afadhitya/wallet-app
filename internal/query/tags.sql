-- name: GetTagByID :one
SELECT * FROM tags WHERE id = ?;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ? COLLATE NOCASE;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES (?) RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;
