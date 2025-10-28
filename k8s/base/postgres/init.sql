CREATE TABLE IF NOT EXISTS raw_windows (
    id bigserial PRIMARY KEY,
    device_id text NOT NULL,
    ts_start timestamptz NOT NULL,
    fs_hz int NOT NULL,
    duration_s real NOT NULL,
    seq bigint NOT NULL,
    label text NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS features(
    id bigserial PRIMARY KEY,
    raw_id bigint NOT NULL REFERENCES raw_windows(id) ON DELETE CASCADE,
    feature_vec jsonb NOT NULL, -- switch to pgvecs later
    label text NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);