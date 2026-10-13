# Remi Jadul

Remi Jadul adalah game kartu Remi online untuk Android dengan nuansa meja kartu modern. Pemain bermain dalam room realtime, menyusun kartu, menurunkan seri atau paralel, mengambil kartu buangan dengan jago, dan mengumpulkan skor selama tiga ronde.

Game state, rules, validasi, giliran, pembagian kartu, joker, scoring, dan hasil akhir dimiliki server. Flutter hanya menampilkan personalized state dan mengirim intent pemain.

## Game overview

- Mendukung 2 sampai 4 pemain dalam satu match.
- Satu match terdiri dari 3 ronde dengan skor kumulatif.
- Kartu dibagikan dari satu deck standar 52 kartu.
- Satu kartu marker dikeluarkan secara rahasia untuk menentukan active joker berdasarkan rank.
- `Seri` adalah minimal 3 kartu satu suit dengan rank berurutan.
- `Paralel` adalah minimal 3 kartu dengan rank sama dan suit berbeda.
- Kombinasi yang sudah diturunkan terkunci dan tidak dapat dipindah atau dipakai ulang.
- Giliran normal terdiri dari draw atau pickup, optional laydown, lalu satu discard.
- Pickup discard memilih kartu dalam discard window dan dua kartu jago. Kartu pilihan wajib langsung menjadi bagian kombinasi; kartu di atasnya ikut masuk ke tangan.
- Menutup ronde memerlukan seluruh kartu tersisa menjadi kombinasi valid setelah satu kartu penutup dibuang.
- Membuang active joker menyebabkan game over dan penalti `-50`.
- Pemenang dan ranking final hanya ditampilkan setelah ronde ketiga.

Detail authoritative berada di [`docs/GAME_RULES.md`](docs/GAME_RULES.md). Keputusan perubahan aturan berada di [`docs/decisions/`](docs/decisions/).

## Product and UX direction

- Gameplay memakai layout meja radial untuk 2, 3, dan 4 pemain.
- Deck berada di tengah meja dan discard pile ditampilkan sebagai ring yang dapat dibuka dalam urutan lengkap.
- Kombinasi seri dan paralel semua pemain tampil sebagai informasi publik.
- Tangan lawan hanya menampilkan ikon jumlah kartu, nama, status online atau offline, dan kombinasi yang sudah diturunkan.
- Tangan sendiri dapat disusun ulang dengan tap atau drag.
- Tap deck digunakan untuk draw.
- Drag kartu ke discard digunakan untuk discard.
- Pickup menggunakan urutan dua kartu jago, lalu satu kartu discard.
- UI tidak memberikan rekomendasi kombinasi. Pemain menentukan pilihannya sendiri dan server memvalidasi.
- Visual menggunakan pseudo-3D tabletop di Flutter dengan perspektif, depth, shadow, dan animasi kartu. Full 3D engine tidak termasuk MVP.

Rencana UI lengkap berada di [`docs/UI_PLAN.md`](docs/UI_PLAN.md). Mock wireframe tersedia di [`docs/examples/`](docs/examples/).

## Technology stack

### Android client

- Flutter and Dart.
- Material 3 sebagai base widget system.
- SVG atau vector card artwork dengan scalable card states.
- WebSocket client untuk personalized realtime state.
- Widget dan contract tests.

### Go backend

- Go pure game engine di `server/internal/game/`.
- Gin untuk REST routing dan WebSocket upgrade boundary.
- `coder/websocket` untuk WebSocket connections.
- GORM dengan PostgreSQL driver untuk repository queries.
- JWT untuk guest session authentication.
- dbmate untuk versioned SQL migrations.
- PostgreSQL untuk guest identity dan completed match results.

Pure game engine tidak boleh mengimpor Gin, WebSocket, GORM, JWT, database, network, clock, atau Flutter packages.

### Protocol

- Versioned JSON v1.
- REST untuk guest identity, public room discovery, dan match history.
- WebSocket untuk room lifecycle, game commands, events, errors, dan resync.
- Personalized complete `game.state` setelah accepted actions.
- Per-recipient event sequence numbers.
- Idempotent `request_id` untuk WebSocket commands.
- Visibility rules yang tidak pernah mengekspos kartu lawan, marker identity, atau closed-deck order.

Protocol source of truth berada di [`docs/protocol/`](docs/protocol/), terutama [`contract.md`](docs/protocol/contract.md) dan [`examples.md`](docs/protocol/examples.md).

## Repository layout

- `AGENTS.md` - instructions dan quality guardrails untuk coding agents.
- `docs/SPEC.md` - product scope dan MVP flows.
- `docs/PRODUCT_BACKLOG.md` - product backlog dan delivery scope.
- `docs/GAME_RULES.md` - authoritative detailed game rules.
- `docs/ARCHITECTURE.md` - system boundaries dan data flow.
- `docs/ROADMAP.md` - implementation phases.
- `docs/decisions/` - durable architecture dan product decisions.
- `docs/coordination/` - PM task packets dan BE/FE handoffs.
- `docs/testing/` - test matrix dan verification scope.
- `docs/DEVELOPMENT.md` - local prerequisites dan agent startup order.
- `.opencode/` - project-local OpenCode commands dan token policy.
- `server/` - Go backend boundary.
- `app/` - Flutter Android client boundary.

## Current implementation status

Repository sedang menyiapkan pure Go engine slice pertama. Task aktif berada di [`docs/coordination/CURRENT_TASK.md`](docs/coordination/CURRENT_TASK.md).

- Specifications, rules, protocol contract, UI plan, dan wireframes sudah terdokumentasi.
- Flutter app shell dan smoke test tersedia.
- Go server package dan verification entrypoints tersedia.
- Card engine, API server, room service, persistence, dan realtime implementation belum selesai.

## Local setup

1. Copy `.env.example` ke `.env` dan isi credential lokal.
2. Install Go, Flutter, Dart, Docker, dan dbmate.
3. Start PostgreSQL saat diperlukan:

```bash
make db-up
```

4. Jalankan focused checks:

```bash
make be-test
make be-fmt
make fe-test
make fe-analyze
make check
```

5. Jalankan migrations setelah migration files tersedia:

```bash
make db-migrate
```

6. Jalankan API setelah `server/cmd/api/main.go` tersedia:

```bash
make be-run
```

## Recommended OpenCode workflow

Gunakan satu scoped task pada satu waktu:

```text
/plan implement the pure deck and deal model
/build implement the approved deck and deal plan
/verify deck and deal
/review deck and deal diff
```

Untuk workflow tiga agent:

```text
/pm define the next pure deck task
/be implement the active deck task
/fe implement the active UI task
/handoff close the active task
```

PM owns `docs/coordination/CURRENT_TASK.md`. BE dan FE tidak boleh bekerja tanpa assigned packet. Hanya satu task boleh `IN_PROGRESS` pada satu waktu.

Untuk simple tasks, review dapat dilewati. Untuk rules, protocol, security, atau cross-package changes, gunakan review.

Stable verification entrypoints tersedia di `Makefile`.

## Current blockers

- Active BE task belum diimplementasikan.
- Executable protocol contract fixtures belum dibuat.
- API entrypoint belum tersedia.
- Migration schema belum dibuat.
- Flutter UI gameplay belum diimplementasikan.

Rule ambiguity baru harus dicatat sebagai proposed decision sebelum implementation.
