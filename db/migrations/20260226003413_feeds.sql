-- +goose Up
-- +goose StatementBegin
create table feeds (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    kind text not null check(kind in ('spotify')),
    created_at datetime not null default current_timestamp,
    last_synced_at datetime,
    unique(user_id, kind)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table feeds;
-- +goose StatementEnd
