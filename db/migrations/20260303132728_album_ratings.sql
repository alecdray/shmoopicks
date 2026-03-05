-- +goose Up
-- +goose StatementBegin
create table album_ratings (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    album_id text not null references albums(id) on delete cascade,
    rating float,
    created_at datetime not null default current_timestamp,
    updated_at datetime,
    unique(user_id, album_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table album_ratings;
-- +goose StatementEnd
