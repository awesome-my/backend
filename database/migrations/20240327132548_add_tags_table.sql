-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tags (
    tag_id SERIAL NOT NULL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(191) NOT NULL,
    slug VARCHAR(191) UNIQUE DEFAULT NULL
);
CREATE TABLE IF NOT EXISTS project_tags (
    project_id INT NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES tags(tag_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS event_tags (
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES tags(tag_id) ON DELETE CASCADE
);
ALTER TABLE projects DROP COLUMN tags;
ALTER TABLE events DROP COLUMN tags;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tags;
DROP TABLE project_tags;
DROP TABLE event_tags;
ALTER TABLE projects ADD COLUMN tags TEXT[] DEFAULT NULL;
ALTER TABLE events ADD COLUMN tags TEXT[] DEFAULT NULL;
-- +goose StatementEnd
