-- name: GetOrCreateUserArtist :exec
INSERT INTO user_artists (user_id, artist_id) VALUES (?, ?)
ON CONFLICT (user_id, artist_id)
DO UPDATE SET user_id = user_id
RETURNING *;
