# Listenly — Roadmap Pengembangan (dari sekarang → production)

> Dokumen ini adalah daftar pekerjaan terurut berdasarkan hasil audit kode saat ini.
> Checklist bisa langsung dicoret (`[x]`) seiring progres.
> Legenda prioritas: 🔴 kritis · 🟠 penting · 🟢 nice-to-have

---

## Ringkasan Status Saat Ini

**Sudah jalan end-to-end (HTTP → gRPC → DB):**
- ✅ Auth: register, login, refresh (rotating), logout, verify/me
- ✅ Room: create (public/private + invite code/token), get, join (public & invite), leave, list public, list my rooms, presence online (Redis)
- ✅ Music: search (Elasticsearch fuzzy), request track (lazy-fetch via media-service), get queue, remove from queue
- ✅ media-service (Python + yt-dlp) sebagai extractor metadata YouTube

**Sudah ada di kode tapi BELUM tersambung ke API (utang teknis):**
- ⚠️ Playback sync (`UpdateState`/`GetState` di Redis) — tidak ada RPC/endpoint
- ⚠️ `MarkAsPlayed` — ada di service + gRPC, tidak ada route gateway
- ⚠️ `RevokeAllUserRefreshTokens` (logout semua perangkat) — ada di repo, belum dipakai
- ⚠️ `GetTrack` by UUID — ada di gRPC, tidak ada route gateway
- ⚠️ `IsOnline` — ada di state repo, belum dipakai

**Belum ada sama sekali:**
- ❌ README, dokumentasi API
- ❌ Dockerfile / docker-compose
- ❌ CI/CD (.github/workflows)
- ❌ Unit/integration test (tidak ada satu pun `*_test.go`)
- ❌ `.env.example`
- ❌ Realtime push ke client (WebSocket/SSE)
- ❌ Message broker (RabbitMQ) — *direncanakan*
- ❌ Observability (metrics, tracing, structured logging terpusat)

---

## FASE 0 — Lengkapi fitur yang setengah jadi 🔴
*Sambungkan yang infrastrukturnya sudah 80% siap. Effort rendah, nilai tinggi.*

- [ ] **Playback sync (inti produk "listen together")**
  - [ ] Tambah RPC di `proto/room`: `UpdatePlayback`, `GetPlayback` (play/pause/seek/current track)
  - [ ] `make proto` untuk regenerate stub
  - [ ] Implementasi handler gRPC room-service memakai `stateRepo.UpdateState/GetState` yang sudah ada
  - [ ] Hanya host yang boleh mengubah state (cek `CheckMembership.isHost`)
  - [ ] Endpoint gateway: `PUT /api/v1/rooms/:uuid/playback`, `GET /api/v1/rooms/:uuid/playback`
- [ ] **Mark as played** — tambah route gateway `POST /api/v1/music/queue/:uuid/played` → `musicClient.MarkAsPlayed`
- [ ] **Auto-advance queue** — saat lagu selesai, set `played` + pindah `current_track_id` ke item berikutnya
- [ ] **Logout semua perangkat** — service `LogoutAll` + endpoint `POST /api/v1/auth/logout-all` (pakai `RevokeAllUserRefreshTokens`)
- [ ] **Get track detail** — route gateway `GET /api/v1/music/tracks/:uuid`
- [ ] **Reorder queue** — RPC + endpoint `PATCH /api/v1/music/queue/:uuid/position` (kolom `position` sudah ada)

---

## FASE 1 — Realtime & RabbitMQ (event-driven) 🔴
*Agar perubahan (lagu baru, play/pause, member join) langsung sampai ke semua client tanpa polling. Ini di mana RabbitMQ masuk.*

- [ ] **Tambah RabbitMQ ke stack**
  - [ ] `pkg/rabbitmq` — koneksi, publisher, consumer, reconnect/retry
  - [ ] Config `RABBITMQ_URL` di tiap service + fallback di `config/env.go`
  - [ ] Tambahkan ke `go.mod`: `github.com/rabbitmq/amqp091-go`
- [ ] **Definisikan topologi exchange/queue** (disarankan topic exchange `listenly.events`)
  - [ ] Routing key: `room.{uuid}.playback`, `room.{uuid}.queue`, `room.{uuid}.presence`
  - [ ] Event: `TrackRequested`, `TrackRemoved`, `PlaybackUpdated`, `MemberJoined`, `MemberLeft`, `TrackPlayed`
- [ ] **Producer**: music-service & room-service publish event setelah perubahan state (queue, playback, presence)
- [ ] **Consumer + fan-out ke client**
  - [ ] Service realtime (bisa di gateway atau service baru) consume event → push via **WebSocket/SSE** ke client di room terkait
  - [ ] Endpoint `GET /api/v1/rooms/:uuid/ws` (WebSocket) atau `/events` (SSE)
- [ ] **Decouple media extraction via RabbitMQ** (opsional tapi natural)
  - [ ] `RequestTrack` publish `MediaExtractRequested` → media-service consume async → publish `MediaExtracted`
  - [ ] Queue item mulai `pending`, jadi `ready` saat metadata siap (status `pending/ready/failed` sudah ada di skema!)
  - [ ] Hilangkan blocking gRPC call saat request track
- [ ] **Reliability RabbitMQ**: dead-letter queue, retry/backoff, idempotent consumer, publisher confirms

---

## FASE 2 — Fitur produk tambahan 🟠
*Butuh migration baru. Lihat `docs/ERD.md` bagian "Fitur yang bisa dibangun".*

- [ ] **Vote/upvote lagu di queue** (crowd-DJ) — migration `queue_votes`, reorder berdasar vote
- [ ] **Chat per room** — migration `messages` + realtime via RabbitMQ (reuse Fase 1)
- [ ] **Favorit / playlist pribadi** — migration `favorites` / `playlists` + `playlist_tracks`
- [ ] **Profil & update user** — endpoint update `full_name`, ganti password
- [ ] **Transfer ownership / delete room** — lengkapi TODO (host saat ini dilarang keluar)
- [ ] **Kick/ban member** — kolom status di `room_members`
- [ ] **Riwayat putar & statistik** — endpoint play history + "lagu terpopuler" (agregasi `queue_items`)
- [ ] **Ekspirasi/regenerasi invite** — kolom `invite_expires_at`
- [ ] **Activity feed** — tabel `room_events` + konsumsi event RabbitMQ

---

## FASE 3 — Kualitas & keandalan 🟠
*Prasyarat sebelum dianggap siap production.*

- [ ] **Testing**
  - [ ] Unit test service layer (auth, room, music) — target coverage awal ~60%
  - [ ] Integration test repository (Postgres/Redis/ES via testcontainers)
  - [ ] Contract test antar gRPC service
  - [ ] `make test` di CI
- [ ] **Validasi & error handling**
  - [ ] Audit mapping gRPC code → HTTP status (beberapa handler masih generik, mis. selalu 400/500)
  - [ ] Standarisasi error envelope lewat `pkg/apperror` + `pkg/response`
  - [ ] Rate limiting di gateway (login, register, request track)
- [ ] **Keamanan**
  - [ ] Pindahkan JWT secret & kredensial DB ke secret manager (jangan di `.env` plaintext)
  - [ ] TLS untuk gRPC antar-service (saat ini `insecure`) — minimal mTLS di prod
  - [ ] CORS policy eksplisit di gateway
  - [ ] Caching hasil `VerifyToken` sejenak agar auth-service tidak jadi bottleneck tiap request
  - [ ] Audit panjang/format input (SQL injection sudah aman via GORM, cek ES query injection)
- [ ] **Pagination konsisten** — gateway saat ini hardcode `page=1, size=20`; teruskan query param

---

## FASE 4 — Infrastruktur & deployment 🔴 (untuk production)

- [ ] **Containerization**
  - [ ] Dockerfile multi-stage per service (gateway, auth, room, music) + media-service (Python)
  - [ ] `docker-compose.yml` untuk dev lokal: Postgres×3, Redis, Elasticsearch, RabbitMQ, semua service
  - [ ] `.dockerignore`, `.env.example` per service
- [ ] **Migrasi database di CI/CD** — jalankan `migrate up` otomatis saat deploy (per DB: `listenly`, `listenly_room`, `listenly_music`)
- [ ] **CI/CD** (`.github/workflows`)
  - [ ] Lint (`golangci-lint`), `go vet`, `buf lint`, `buf breaking`
  - [ ] Test + build + push image ke registry
  - [ ] Deploy otomatis (staging → prod)
- [ ] **Orkestrasi** — Kubernetes manifest/Helm atau Nomad; health/readiness probe (`/health` sudah ada di gateway, tambahkan gRPC health check di service lain)
- [ ] **Konfigurasi lingkungan** — pisahkan config dev/staging/prod; hostname service via service discovery (bukan hardcode)

---

## FASE 5 — Observability & operasional 🟠

- [ ] **Logging terstruktur terpusat** — `slog` sudah dipakai; kirim ke Loki/ELK, tambahkan request ID & trace ID
- [ ] **Metrics** — Prometheus (RED metrics per endpoint, lag consumer RabbitMQ, latency gRPC)
- [ ] **Tracing** — OpenTelemetry (dependency OTel sudah ada di `go.mod`!) span lintas service + RabbitMQ
- [ ] **Dashboard & alerting** — Grafana + alert (error rate, latency p99, antrian RabbitMQ menumpuk)
- [ ] **Backup & retention** — backup Postgres terjadwal, kebijakan retensi data queue/history
- [ ] **Graceful shutdown** — tutup koneksi gRPC/Redis/RabbitMQ dengan rapi saat SIGTERM

---

## FASE 6 — Dokumentasi & developer experience 🟢

- [ ] **README.md** — arsitektur, cara run lokal, diagram (link `docs/ERD.md`)
- [ ] **Dokumentasi API** — OpenAPI/Swagger untuk endpoint gateway
- [ ] **CONTRIBUTING.md** — konvensi commit, branch, PR
- [ ] **Postman/Bruno collection** untuk testing manual
- [ ] **ADR (Architecture Decision Records)** — catat keputusan: kenapa gRPC, kenapa 3 DB terpisah, kenapa RabbitMQ

---

## Urutan Rekomendasi (quick wins dulu)

1. **Fase 0** — sambungkan playback sync + mark-as-played (inti produk, infra siap)
2. **Fase 4 (sebagian)** — docker-compose dev agar semua dependency (termasuk **RabbitMQ**) mudah dijalankan
3. **Fase 1** — RabbitMQ + realtime WebSocket (ini yang membuat "listen together" terasa hidup)
4. **Fase 3** — testing + keamanan sebelum production
5. **Fase 4 lengkap + Fase 5** — deployment & observability
6. **Fase 2 & 6** — fitur tambahan & dokumentasi, berjalan paralel
```
