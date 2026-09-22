CREATE TABLE rooms (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    host_id BIGINT NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    invite_code VARCHAR(6) UNIQUE,
    invite_token VARCHAR(64) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rooms_uuid ON rooms(uuid);
CREATE INDEX idx_rooms_host_id ON rooms(host_id);
CREATE INDEX idx_rooms_invite_code ON rooms(invite_code);
CREATE INDEX idx_rooms_invite_token ON rooms(invite_token);
CREATE INDEX idx_rooms_visibility ON rooms(visibility);