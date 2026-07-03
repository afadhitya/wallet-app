-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = ?;

-- name: GetCategoryByName :one
SELECT * FROM categories WHERE name = ? COLLATE NOCASE AND is_archived = 0;

-- name: ListCategories :many
SELECT * FROM categories WHERE is_archived = 0 ORDER BY sort_order, name;

-- name: ListAllCategories :many
SELECT * FROM categories ORDER BY sort_order, name;

-- name: CreateCategory :one
INSERT INTO categories (name, parent_id, type, icon) VALUES (?, ?, ?, ?) RETURNING *;

-- name: UpdateCategory :one
UPDATE categories SET name = ?, icon = ?, updated_at = datetime('now') WHERE id = ? RETURNING *;

-- name: ArchiveCategory :exec
UPDATE categories SET is_archived = 1, updated_at = datetime('now') WHERE id = ?;

-- name: GetCategorySuggestions :many
SELECT * FROM categories WHERE name LIKE ? COLLATE NOCASE AND is_archived = 0 LIMIT 5;
