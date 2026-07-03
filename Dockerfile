FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY docs ./docs
COPY version.go ./

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o seed ./cmd/seed

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /app/api /api
COPY --from=build /app/seed /seed
COPY --from=build /app/internal/infra/db/migrations /app/internal/infra/db/migrations

EXPOSE 8080

CMD ["/api"]
