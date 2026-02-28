-- name: GetOrCreateUserRelease :one
INSERT INTO user_releases (id, user_id, release_id) VALUES (?, ?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET user_id = user_id
RETURNING *;

-- name: GetUserReleases :many
SELECT sqlc.embed(user_releases), sqlc.embed(releases) FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ?;
