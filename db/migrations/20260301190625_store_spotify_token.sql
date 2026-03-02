-- +goose Up
-- +goose StatementBegin
alter table users add column spotify_refresh_token text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users drop column spotify_refresh_token;
-- +goose StatementEnd
