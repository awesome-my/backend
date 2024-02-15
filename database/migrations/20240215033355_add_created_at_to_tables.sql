-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT now();
ALTER TABLE projects ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN created_at;
ALTER TABLE projects DROP COLUMN created_at;
-- +goose StatementEnd
