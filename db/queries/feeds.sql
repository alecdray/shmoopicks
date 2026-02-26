-- name: CreateFeed :one
insert into feeds (user_id, kind)
values (?, ?)
returning *;

-- name: UpsertFeed :one
INSERT INTO feeds (id, user_id, kind)
VALUES (?, ?, ?)
ON CONFLICT (user_id, kind) DO UPDATE SET
    user_id = excluded.user_id,
    kind = excluded.kind
RETURNING *;

-- name: GetFeedsByUserId :many
select * from feeds where user_id = ?;

-- name: GetStaleFeedsBatch :many
SELECT * FROM feeds
WHERE last_synced_at < datetime('now', ?)
OR last_synced_at IS NULL
ORDER BY last_synced_at ASC
LIMIT 10;
