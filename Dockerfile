# ---------- Stage 1: Build ----------
FROM golang:1.24.3 AS builder

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum files first (for caching dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
RUN go build -o eth-watcher ./cmd/app

# ---------- Stage 2: Run ----------
FROM debian:bookworm-slim

# Install CA certificates to enable TLS verification (fixes x509 errors)
# And clean up apt cache to keep image size small
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/eth-watcher .

EXPOSE 8080

# Run the binary
CMD ["./eth-watcher"]