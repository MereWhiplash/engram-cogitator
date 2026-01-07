# Build stage
FROM golang:1.24-bookworm AS builder

# Install build dependencies for CGO (required for sqlite-vec)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Tidy modules (in case go.sum is stale)
RUN go mod tidy

# Build with CGO enabled for sqlite-vec
ENV CGO_ENABLED=1
RUN go build -o /ec-api ./cmd/api
RUN go build -o /ec-shim ./cmd/shim
RUN go build -o /engram-cogitator ./cmd/server

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

# Copy binaries from builder
COPY --from=builder /ec-api /ec-api
COPY --from=builder /ec-shim /ec-shim
COPY --from=builder /engram-cogitator /usr/local/bin/engram-cogitator

# Create data directory
RUN mkdir -p /data

ENTRYPOINT ["/ec-api"]
CMD ["--addr", ":8080"]
