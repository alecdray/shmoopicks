-- name: UpsertAlbumRating :one
INSERT INTO album_ratings (id, user_id, album_id, rating) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, album_id)
DO UPDATE SET rating = COALESCE(EXCLUDED.rating, rating), updated_at = current_timestamp
RETURNING *;

-- name: GetUserAlbumRatings :many
select * from album_ratings
where user_id = ?;

-- name: GetUserAlbumRating :one
select * from album_ratings
where user_id = ?
and album_id = ?;

-- name: GetUserAlbumRatingById :one
select * from album_ratings
where id = ?;
