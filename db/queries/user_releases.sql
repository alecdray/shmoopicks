-- name: UpsertUserRelease :one
INSERT INTO user_releases (id, user_id, release_id, added_at) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET added_at = COALESCE(EXCLUDED.added_at, added_at)
RETURNING *;

-- name: GetUserReleases :many
SELECT sqlc.embed(user_releases), sqlc.embed(releases) FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ?;
