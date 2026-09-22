CREATE TABLE queue_items (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    room_id BIGINT NOT NULL,
    track_id BIGINT NOT NULL REFERENCES tracks(id),
    requested_by BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'failed', 'played')),
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_queue_items_uuid ON queue_items(uuid);
CREATE INDEX idx_queue_items_room_id ON queue_items(room_id);
CREATE INDEX idx_queue_items_status ON queue_items(status);