run:
	go run ./cmd/api

migrate-up:
	docker compose run --rm migrate
