-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN github_email TYPE TEXT;
ALTER TABLE users ALTER COLUMN github_email DROP NOT NULL;
ALTER TABLE users ADD COLUMN google_email TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN google_email;
ALTER TABLE users ALTER COLUMN github_email TYPE VARCHAR(64);
ALTER TABLE users ALTER COLUMN github_email SET NOT NULL;
-- +goose StatementEnd
