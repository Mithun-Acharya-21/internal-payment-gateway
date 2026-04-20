# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a statically linked binary — no runtime dependencies
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -X main.version=$(git describe --tags --always)" \
    -o /bin/gateway ./cmd/server

# ── Stage 2: Production image ─────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS production

# Run as non-root user for security
USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /bin/gateway /gateway

EXPOSE 8080

ENTRYPOINT ["/gateway"]
