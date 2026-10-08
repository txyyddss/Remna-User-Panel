-- Retain bounded diagnostic attempts only; proxy configurations stay transient.
CREATE TABLE host_connectivity_attempts (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    host_uuid TEXT,
    config_hash TEXT NOT NULL CHECK (length(config_hash)=64),
    remnawave_user_id INTEGER NOT NULL CHECK (remnawave_user_id>0),
    trigger TEXT NOT NULL CHECK (trigger IN ('manual','scheduled')),
    started_at TEXT NOT NULL CHECK (length(started_at)=30),
    finished_at TEXT CHECK (finished_at IS NULL OR (length(finished_at)=30 AND finished_at>=started_at)),
    status TEXT NOT NULL CHECK (status IN ('running','connected','failed','unsupported','error','interrupted')),
    latency_ms REAL CHECK (latency_ms IS NULL OR latency_ms>=0),
    http_status INTEGER CHECK (http_status IS NULL OR http_status BETWEEN 100 AND 599),
    error_code TEXT NOT NULL DEFAULT '',
    CHECK ((status='running' AND finished_at IS NULL AND latency_ms IS NULL AND http_status IS NULL AND error_code='')
        OR (status<>'running' AND finished_at IS NOT NULL))
);

CREATE INDEX host_connectivity_attempts_started_idx ON host_connectivity_attempts(started_at DESC,id DESC);
CREATE INDEX host_connectivity_attempts_host_idx ON host_connectivity_attempts(host_uuid,started_at DESC,id DESC);
CREATE INDEX host_connectivity_attempts_latest_idx ON host_connectivity_attempts(config_hash,host_uuid,started_at DESC,id DESC);
