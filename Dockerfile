FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -X main.version=$(git describe --tags --always)" \
    -o /bin/gateway ./cmd/server

FROM gcr.io/distroless/static-debian12 AS production

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /bin/gateway /gateway

EXPOSE 8080

ENTRYPOINT ["/gateway"]
