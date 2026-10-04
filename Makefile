.PHONY: proto migrate-up migrate-down migrate-create run-gateway run-auth run-media media-setup

# ── Proto ──────────────────────────────
proto:
	cd proto && buf lint && buf generate

proto-breaking:
	cd proto && buf breaking --against '.git#branch=main'

# ── Migration ──────────────────────────
DB_URL=postgres://postgres:postgres@localhost:54322/listenly?sslmode=disable

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-force:
	@read -p "Version: " version; \
	migrate -path migrations -database "$(DB_URL)" force $$version

# ── Run Services ───────────────────────
run-gateway:
	go run ./cmd/gateway

run-auth:
	go run ./cmd/auth-service

run-room:
	go run ./cmd/room-service

run-music:
	go run ./cmd/music-service

# ── Media Service (Python) ─────────────
MEDIA_DIR=media-service
VENV_PY=$(MEDIA_DIR)/venv/bin/python
VENV_PIP=$(MEDIA_DIR)/venv/bin/pip

run-media:
	cd $(MEDIA_DIR)/app && ../venv/bin/python main.py

media-setup:
	cd $(MEDIA_DIR) && python3.12 -m venv venv && \
		venv/bin/pip install --upgrade pip && \
		venv/bin/pip install -e .


# ── Dev ────────────────────────────────
tidy:
	go mod tidy

test:
	go test ./... -v