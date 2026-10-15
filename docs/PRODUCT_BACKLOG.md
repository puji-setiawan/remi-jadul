# Remi Jadul — Product Backlog

Dokumen ini berisi backlog dan scope produk. Aturan permainan authoritative berada di [`GAME_RULES.md`](GAME_RULES.md), bukan di backlog ini.

## MVP goal

Game kartu Remi online Android untuk 2–4 pemain dengan room publik/privat, permainan realtime, hasil ronde, dan ranking.

## MVP capabilities

- Guest identity dengan nickname.
- Room publik yang dapat ditemukan selama slot tersedia.
- Room privat dengan kode atau undangan.
- Room berkapasitas 2–4 pemain.
- Host memulai game saat jumlah pemain yang dipilih tersedia.
- Server membagikan kartu, menentukan joker, memvalidasi aksi, mengatur giliran, menghitung skor, dan mengakhiri ronde.
- Match terdiri dari tiga ronde dengan skor kumulatif dan pemenang akhir.
- Gameplay menggunakan presentasi pseudo-3D tabletop di Flutter dengan perspektif, depth, dan animasi kartu.
- Draw dari closed deck dan discard pickup sesuai `GAME_RULES.md`.
- Personalized state projection tanpa membocorkan kartu lawan.
- Reconnect dasar ke room aktif.
- Penyimpanan guest/profile dan hasil ronde di PostgreSQL.

## Technology baseline

| Area | Technology | Responsibility |
|---|---|---|
| Android client | Flutter / Dart | Lobby, room, cards, actions, score, reconnect UX |
| Backend | Go | Game engine, rules, room, validation, scoring |
| Realtime | Versioned JSON over WebSocket | Active-game commands and events |
| Ordinary API | REST/HTTP | Guest identity, profile, room discovery, history |
| Database | PostgreSQL | Guest/profile data and completed results |
| Local development | Docker Compose | PostgreSQL and future local services |

## Explicit non-goals

- Redis and multi-instance active-room state.
- Active-round recovery after server restart.
- OAuth, payments, chat, spectators, tournaments, and rating matchmaking.

## Delivery backlog

### Phase 1 — Pure Go game engine

- [x] Card model and physical card IDs.
- [x] Standard deck, shuffle, deal, and joker marker.
- [x] Seri and paralel validation.
- [x] Joker representation validation.
- [x] Table combination locking.
- [x] Turn state and legal action validation.
- [x] Closed-deck draw and discard pickup.
- [x] Close, game-over, and deterministic scoring.
- [x] Table-driven unit tests for valid, invalid, and boundary cases.

### Phase 2 — Room and guest identity

- [x] Guest nickname identity.
- [x] Public/private room lifecycle.
- [x] Host start and 2–4 player validation.
- [x] In-memory active rooms.

### Phase 3 — REST and WebSocket protocol

- [x] Versioned protocol in `docs/protocol/`.
- [x] Personalized state projections.
- [x] Action errors and request IDs.
- [x] Event sequence/order handling.
- [x] Reconnect and stale-event behavior.
- [x] Visibility and contract tests.

### Phase 4 — PostgreSQL persistence

- [x] Versioned migrations.
- [x] Guest/profile repository.
- [x] Completed match and round result repository.
- [x] Persistence integration tests.

### Phase 5 — Flutter Android client

- [ ] Guest entry and lobby.
- [ ] Public/private room create and join.
- [ ] Room waiting state and host start.
- [ ] Personalized game table and hand.
- [ ] Pseudo-3D tabletop presentation with card depth, perspective, and movement animation.
- [ ] Draw, discard, laydown, discard pickup, and close interactions.
- [ ] Score, game-over, ranking, loading, error, and reconnect states.
- [ ] Android emulator acceptance flow.

### Phase 6 — MVP hardening

- [ ] Dockerized backend workflow.
- [ ] CI gates for Go, protocol, and Flutter checks.
- [ ] Structured logs without hidden-card leakage.
- [ ] Go race detection.
- [ ] Security and reconnect review.
- [ ] Deployment and release checklist.
