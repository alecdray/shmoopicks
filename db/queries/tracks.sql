-- name: CreateTrack :exec
INSERT INTO tracks (id, spotify_id, title) VALUES (?, ?, ?);

-- name: GetOrCreateTrack :one
INSERT INTO tracks (id, spotify_id, title) VALUES (?, ?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET spotify_id = spotify_id
RETURNING *;

-- name: GetTrack :one
SELECT * FROM tracks WHERE id = ?;

-- name: GetTrackBySpotifyId :one
SELECT * FROM tracks WHERE spotify_id = ?;
