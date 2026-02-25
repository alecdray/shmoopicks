-- +goose Up
-- +goose StatementBegin

create table users (
    id text primary key,
    spotify_id text not null unique,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
);

create table artists (
    id text primary key,
    spotify_id text not null unique,
    name text not null,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
);

create table albums (
    id text primary key,
    spotify_id text not null unique,
    title text not null,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
);

create table tracks (
    id text primary key,
    spotify_id text not null unique,
    title text not null,
    album_id text not null references albums(id) on delete cascade,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
);

create table album_artists (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    artist_id text not null references artists(id) on delete cascade,
    unique(album_id, artist_id)
);

create table album_tracks (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    unique(album_id, track_id)
);

create table releases (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    format text not null check(format in ('digital', 'vinyl', 'cd', 'cassette')),
    created_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(album_id, format)
);

create table user_releases (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    release_id text not null references releases(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(user_id, release_id)
);

create table user_tracks (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(user_id, track_id)
);

create table user_artists (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    artist_id text not null references artists(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(user_id, artist_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table user_artists;
drop table user_tracks;
drop table user_releases;
drop table releases;
drop table album_tracks;
drop table album_artists;
drop table albums;
drop table tracks;
drop table artists;
drop table users;
-- +goose StatementEnd
