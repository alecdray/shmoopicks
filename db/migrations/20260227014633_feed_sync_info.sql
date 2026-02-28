-- +goose Up
-- +goose StatementBegin
alter table feeds drop column last_synced_at;
alter table feeds add column last_sync_completed_at datetime;
alter table feeds add column last_sync_started_at datetime;
alter table feeds add column last_sync_status text default 'none' check(last_sync_status in ('none', 'success', 'failure', 'pending'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table feeds drop column last_sync_status;
alter table feeds drop column last_sync_started_at;
alter table feeds drop column last_sync_completed_at;
alter table feeds add column last_synced_at datetime;
-- +goose StatementEnd
