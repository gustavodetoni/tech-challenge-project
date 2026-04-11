FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api ./cmd/api

RUN GOBIN=/go/bin go install github.com/pressly/goose/v3/cmd/goose@v3.25.0

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /app/api /api
COPY --from=build /go/bin/goose /goose
COPY --from=build /app/internal/infra/db/migrations /app/internal/infra/db/migrations

EXPOSE 8080

CMD ["/api"]
