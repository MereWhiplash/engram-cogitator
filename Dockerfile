# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies for CGO (required for sqlite-vec)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for sqlite-vec
ENV CGO_ENABLED=1
RUN go build -o /engram-cogitator ./cmd/server

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /engram-cogitator /usr/local/bin/engram-cogitator

# Create data directory
RUN mkdir -p /data

ENTRYPOINT ["engram-cogitator"]
CMD ["--db-path", "/data/memory.db", "--ollama-url", "http://ollama:11434"]
