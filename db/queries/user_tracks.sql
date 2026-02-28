-- name: GetOrCreateUserTrack :exec
INSERT INTO user_tracks (id, user_id, track_id) VALUES (?, ?, ?)
ON CONFLICT (user_id, track_id)
DO UPDATE SET user_id = user_id
RETURNING *;

-- name: GetUserTracks :many
SELECT sqlc.embed(user_tracks), sqlc.embed(tracks) FROM user_tracks
JOIN tracks ON user_tracks.track_id = tracks.id
WHERE user_id = ?;
