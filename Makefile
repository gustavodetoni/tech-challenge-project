run:
	go run ./cmd/api

dev:
	air -c .air.toml

build:
	mkdir -p bin
	go build -o ./bin/api ./cmd/api

tests:
	mkdir -p /tmp/go-tmp /tmp/go-build-cache /tmp/go-mod-cache
	env GOCACHE=/tmp/go-build-cache GOTMPDIR=/tmp/go-tmp \
		go test ./... -coverpkg=./... -covermode=atomic -coverprofile=coverage.out -count=1
	env GOCACHE=/tmp/go-build-cache GOTMPDIR=/tmp/go-tmp \
		go tool cover -func=coverage.out | grep total:

swagger:
	mkdir -p /tmp/go-tmp /tmp/go-build-cache /tmp/go-mod-cache
	env GOCACHE=/tmp/go-build-cache GOTMPDIR=/tmp/go-tmp \
		go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs

migrate-up:
	docker compose run --rm migrate

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

fmt:
	goimports -w .
