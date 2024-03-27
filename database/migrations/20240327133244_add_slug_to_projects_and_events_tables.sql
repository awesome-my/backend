-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN slug VARCHAR(191) UNIQUE DEFAULT NULL;
ALTER TABLE events ADD COLUMN slug VARCHAR(191) UNIQUE DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN slug;
ALTER TABLE events DROP COLUMN slug;
-- +goose StatementEnd
