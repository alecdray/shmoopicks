-- +goose Up
-- +goose StatementBegin
alter table tracks drop column album_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table tracks add column album_id text not null references albums(id) on delete cascade;
-- +goose StatementEnd
