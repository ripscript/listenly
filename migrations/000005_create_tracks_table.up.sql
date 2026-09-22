CREATE TABLE tracks (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    youtube_video_id VARCHAR(20) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255),
    duration_seconds INT NOT NULL DEFAULT 0,
    thumbnail_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tracks_uuid ON tracks(uuid);
CREATE INDEX idx_tracks_youtube_video_id ON tracks(youtube_video_id);
CREATE INDEX idx_tracks_title ON tracks(title);