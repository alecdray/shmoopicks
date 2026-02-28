-- name: GetOrCreateAlbumArtist :one
INSERT INTO album_artists (album_id, artist_id) VALUES (?, ?)
ON CONFLICT (album_id, artist_id)
DO UPDATE SET album_id = album_id
RETURNING *;

-- name: GetAlbumArtistByAlbumId :many
SELECT album_artists.album_id, sqlc.embed(artists) FROM album_artists
JOIN artists ON album_artists.artist_id = artists.id
WHERE album_id = ?;

-- name: GetAlbumArtistsByAlbumIds :many
SELECT album_artists.album_id, sqlc.embed(artists) FROM album_artists
JOIN artists ON album_artists.artist_id = artists.id
WHERE album_id IN (sqlc.slice('album_ids'));
