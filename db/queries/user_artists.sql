-- name: GetOrCreateUserArtist :exec
INSERT INTO user_artists (id, user_id, artist_id) VALUES (?, ?, ?)
ON CONFLICT (user_id, artist_id)
DO UPDATE SET user_id = user_id
RETURNING *;

-- name: GetUserArtists :many
SELECT sqlc.embed(user_artists), sqlc.embed(artists) FROM user_artists
JOIN artists ON user_artists.artist_id = artists.id
WHERE user_id = ?;
