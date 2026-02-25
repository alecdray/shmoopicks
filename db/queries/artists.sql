-- name: CreateArtist :exec
INSERT INTO artists (id, spotify_id, name) VALUES (?, ?, ?);

-- name: GetOrCreateArtist :exec
INSERT INTO artists (id, spotify_id, name) VALUES (?, ?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET spotify_id = spotify_id
RETURNING *;

-- name: GetArtist :one
SELECT * FROM artists WHERE id = ?;

-- name: GetArtistBySpotifyId :one
SELECT * FROM artists WHERE spotify_id = ?;
