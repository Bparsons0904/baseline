-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY
);

-- +migrate Down
DROP TABLE IF EXISTS users;
