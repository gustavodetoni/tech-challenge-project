run:
	go run ./cmd/api

dev:
	air -c .air.toml

build:
	mkdir -p bin
	go build -o ./bin/api ./cmd/api

tests:
	mkdir -p /tmp/go-tmp /tmp/go-build-cache /tmp/go-mod-cache
	env GOCACHE=/tmp/go-build-cache GOTMPDIR=/tmp/go-tmp GOMODCACHE=/tmp/go-mod-cache \
		go test ./... -coverpkg=./... -coverprofile=/tmp/coverage.out -count=1
	go tool cover -func=/tmp/coverage.out | grep total:

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs

migrate-up:
	docker compose run --rm migrate

lint:
	mkdir -p /tmp/go-tmp /tmp/go-build-cache /tmp/go-mod-cache /tmp/xdg-cache /tmp/golangci-lint-cache
	env GOCACHE=/tmp/go-build-cache GOTMPDIR=/tmp/go-tmp GOMODCACHE=/tmp/go-mod-cache XDG_CACHE_HOME=/tmp/xdg-cache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache \
		golangci-lint run
