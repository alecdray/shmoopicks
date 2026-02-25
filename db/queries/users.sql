-- name: CreateUser :exec
INSERT INTO users (id, spotify_id) VALUES (?, ?);

-- name: GetOrCreateUser :exec
INSERT INTO users (id, spotify_id) VALUES (?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET spotify_id = spotify_id
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserBySpotifyId :one
SELECT * FROM users WHERE spotify_id = ?;
