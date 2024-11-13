-- name: GetState :one
SELECT * FROM states
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateState :exec
INSERT INTO states (user_id, state, data, meta, created_at) VALUES ($1, $2, $3, $4, now());
