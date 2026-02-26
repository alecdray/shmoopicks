-- name: CreateUser :one
INSERT INTO users (id, spotify_id) VALUES (?, ?)
RETURNING *;

-- name: UpsertSpotifyUser :one
INSERT INTO users (id, spotify_id) VALUES (?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET spotify_id = EXCLUDED.spotify_id
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserBySpotifyId :one
SELECT * FROM users WHERE spotify_id = ?;
