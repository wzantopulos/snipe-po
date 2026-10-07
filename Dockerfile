# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies for CGO (sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled for sqlite3
ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o snipe-po .

# Runtime stage - minimal image with sqlite3
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Create non-root user
RUN adduser -D -u 1000 snipepo

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/snipe-po .

# Copy settings example
COPY settings.example.yaml .

# Create data directory
RUN mkdir -p /app/data && chown -R snipepo:snipepo /app

# Switch to non-root user
USER snipepo

# Expose port
EXPOSE 80

# Set volume for persistent data
VOLUME ["/app/data"]

# Entry point
ENTRYPOINT ["./snipe-po"]
CMD ["serve", "--port", "80"]
