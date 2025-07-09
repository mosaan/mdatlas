# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1001 mdatlas && \
    adduser -D -u 1001 -G mdatlas mdatlas

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/mdatlas /usr/local/bin/mdatlas

# Create directory for documents
RUN mkdir -p /docs && chown mdatlas:mdatlas /docs

# Switch to non-root user
USER mdatlas

# Set default base directory
ENV BASE_DIR="/docs"

# Expose port (if needed for future HTTP endpoint)
EXPOSE 8080

# Default command
ENTRYPOINT ["mdatlas"]
CMD ["--mcp-server", "--base-dir", "/docs"]