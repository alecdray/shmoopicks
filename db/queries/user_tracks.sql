-- name: GetOrCreateUserTrack :exec
INSERT INTO user_tracks (user_id, track_id) VALUES (?, ?)
ON CONFLICT (user_id, track_id)
DO UPDATE SET user_id = user_id
RETURNING *;
