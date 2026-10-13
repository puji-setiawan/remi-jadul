.PHONY: check be-test be-fmt be-run db-up db-down db-migrate db-rollback db-status fe-test fe-analyze

-include .env
export

DBMATE ?= dbmate
MIGRATIONS_DIR ?= server/migrations

check: be-fmt be-test fe-test fe-analyze

be-test:
	cd server && go test ./...

be-fmt:
	cd server && test -z "$$(gofmt -l .)"

be-run:
	@test -f server/cmd/api/main.go || (echo "BE API entrypoint is not implemented: server/cmd/api/main.go"; exit 1)
	cd server && go run ./cmd/api

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-migrate:
	@command -v $(DBMATE) >/dev/null 2>&1 || (echo "dbmate is required; install it before running migrations"; exit 1)
	@test -d $(MIGRATIONS_DIR) || (echo "Migration directory is missing: $(MIGRATIONS_DIR)"; exit 1)
	$(DBMATE) --url "$(DATABASE_URL)" --migrations-dir "$(MIGRATIONS_DIR)" up

db-rollback:
	@command -v $(DBMATE) >/dev/null 2>&1 || (echo "dbmate is required; install it before rolling back migrations"; exit 1)
	@test -d $(MIGRATIONS_DIR) || (echo "Migration directory is missing: $(MIGRATIONS_DIR)"; exit 1)
	$(DBMATE) --url "$(DATABASE_URL)" --migrations-dir "$(MIGRATIONS_DIR)" rollback

db-status:
	@command -v $(DBMATE) >/dev/null 2>&1 || (echo "dbmate is required; install it before checking migration status"; exit 1)
	@test -d $(MIGRATIONS_DIR) || (echo "Migration directory is missing: $(MIGRATIONS_DIR)"; exit 1)
	$(DBMATE) --url "$(DATABASE_URL)" --migrations-dir "$(MIGRATIONS_DIR)" status

fe-test:
	cd app && flutter test

fe-analyze:
	cd app && flutter analyze
