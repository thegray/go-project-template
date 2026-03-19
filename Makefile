.PHONY: run test build migrate-create migrate-up migrate-down migrate-drop migrate-version migrate-seed

run:
	go run ./cmd/server

build:
	go build ./...

test:
	go test ./...

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-drop:
	go run ./cmd/migrate drop

migrate-version:
	go run ./cmd/migrate version

migrate-seed:
	go run ./cmd/migrate seed

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_users_table" && exit 1)
	go run ./cmd/migrate create $(name)
