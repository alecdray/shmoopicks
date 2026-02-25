-- name: GetOrCreateAlbumTracks :exec
INSERT INTO album_tracks (album_id, track_id) VALUES (?, ?)
ON CONFLICT (album_id, track_id)
DO UPDATE SET album_id = album_id
RETURNING *;
