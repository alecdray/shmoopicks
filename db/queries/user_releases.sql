-- name: GetOrCreateUserRelease :exec
INSERT INTO user_releases (user_id, release_id) VALUES (?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET user_id = user_id
RETURNING *;
