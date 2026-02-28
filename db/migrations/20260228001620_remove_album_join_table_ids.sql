-- +goose Up
CREATE TABLE album_artists_new (
    album_id text not null references albums(id) on delete cascade,
    artist_id text not null references artists(id) on delete cascade,
    unique(album_id, artist_id)
);
INSERT INTO album_artists_new SELECT album_id, artist_id FROM album_artists;
DROP TABLE album_artists;
ALTER TABLE album_artists_new RENAME TO album_artists;

CREATE TABLE album_tracks_new (
    album_id text not null references albums(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    unique(album_id, track_id)
);
INSERT INTO album_tracks_new SELECT album_id, track_id FROM album_tracks;
DROP TABLE album_tracks;
ALTER TABLE album_tracks_new RENAME TO album_tracks;

-- +goose Down
CREATE TABLE album_artists_old (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    artist_id text not null references artists(id) on delete cascade,
    unique(album_id, artist_id)
);
INSERT INTO album_artists_old SELECT hex(randomblob(16)), album_id, artist_id FROM album_artists;
DROP TABLE album_artists;
ALTER TABLE album_artists_old RENAME TO album_artists;

CREATE TABLE album_tracks_old (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    unique(album_id, track_id)
);
INSERT INTO album_tracks_old SELECT hex(randomblob(16)), album_id, track_id FROM album_tracks;
DROP TABLE album_tracks;
ALTER TABLE album_tracks_old RENAME TO album_tracks;
