run:
	go run ./cmd/api

build:
	mkdir -p bin
	go build -o ./bin/api ./cmd/api

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs

dev:
	air -c .air.toml

migrate-up:
	docker compose run --rm migrate
