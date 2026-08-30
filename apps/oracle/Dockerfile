# ── Build stage ──────────────────────────────────────────────────────
FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    liboqs-dev \
    gcc \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o /app/bin/main ./cmd/app/main.go

# ── Runtime stage ────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    liboqs0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/main .
COPY .env .env

EXPOSE 9000

CMD ["./main"]
