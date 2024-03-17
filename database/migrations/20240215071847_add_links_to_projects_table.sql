-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN repository VARCHAR(191) DEFAULT NULL;
ALTER TABLE projects ADD COLUMN website VARCHAR(191) DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN repository;
ALTER TABLE projects DROP COLUMN website;
-- +goose StatementEnd
