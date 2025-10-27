CREATE TABLE IF NOT EXISTS healthcheck (
    id serial PRIMARY KEY,
    created_at timestamptz DEFAULT now()
);