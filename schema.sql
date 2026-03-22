CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  password TEXT,
  createdAt TIMESTAMPTZ DEFAULT now()
);

