-- +goose Up
CREATE TABLE note (
  id uuid,
  title text,
  content text,
  created_at timestamp,
  updated_at timestamp,
  PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE note;