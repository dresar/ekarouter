FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.Version=1.0.0" -o /app/ekarouter ./cmd/ekarouter

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -g 10001 -S ekarouter && \
    adduser -u 10001 -S ekarouter -G ekarouter && \
    mkdir -p /app/data /app/migrations && \
    chown -R ekarouter:ekarouter /app

WORKDIR /app

COPY --from=builder /app/ekarouter /app/ekarouter
COPY migrations/ /app/migrations/

RUN chown -R ekarouter:ekarouter /app

USER ekarouter

EXPOSE 8080

VOLUME ["/app/data"]

HEALTHCHECK --interval=20s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/ekarouter"]
CMD ["serve", "-host", "0.0.0.0", "-port", "8080", "-db", "/app/data/ekarouter.db", "-migrations", "/app/migrations"]
