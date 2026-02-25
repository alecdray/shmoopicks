-- name: GetOrCreateAlbumArtist :exec
INSERT INTO album_artists (album_id, artist_id) VALUES (?, ?)
ON CONFLICT (album_id, artist_id)
DO UPDATE SET album_id = album_id
RETURNING *;
