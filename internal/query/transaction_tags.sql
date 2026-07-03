-- name: AddTransactionTag :exec
INSERT OR IGNORE INTO transaction_tags (transaction_id, tag_id) VALUES (?, ?);

-- name: RemoveTransactionTag :exec
DELETE FROM transaction_tags WHERE transaction_id = ? AND tag_id = ?;

-- name: ListTransactionTags :many
SELECT t.* FROM tags t JOIN transaction_tags tt ON tt.tag_id = t.id WHERE tt.transaction_id = ? ORDER BY t.name;

-- name: DeleteTransactionTags :exec
DELETE FROM transaction_tags WHERE transaction_id = ?;
