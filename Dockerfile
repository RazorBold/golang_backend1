# ── Stage 1: build ───────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build server binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /out/server ./cmd/server

# Build worker binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/worker ./cmd/worker

# ── Stage 2: runtime ─────────────────────────────────────────────
FROM scratch

# Copy timezone data and CA certs from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /out/server /server
COPY --from=builder /out/worker /worker

# Non-root user (numeric UID works with scratch)
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]
