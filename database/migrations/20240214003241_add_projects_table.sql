-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS projects (
    project_id SERIAL NOT NULL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(191) NOT NULL,
    description TEXT NOT NULL,
    tags TEXT[] DEFAULT NULL,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE projects;
-- +goose StatementEnd
