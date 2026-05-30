FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app && \
    mkdir -p /app/internal/infrastructure/db/migrations /var/log/paris-api /app/tmp/logs && \
    chown -R app:app /app /var/log/paris-api

WORKDIR /app

COPY --from=builder /out/server /usr/local/bin/server
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY --from=builder /src/internal/infrastructure/db/migrations /app/internal/infrastructure/db/migrations

USER app

EXPOSE 9000

ENTRYPOINT ["/usr/local/bin/server"]
