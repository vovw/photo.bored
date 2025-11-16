# Multi-stage build for minimal image size
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./
COPY index.html ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o app .

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and static files from builder
COPY --from=builder /build/app .
COPY --from=builder /build/index.html .

# Create directories for data and uploads
RUN mkdir -p /app/data /app/uploads && \
    chmod 755 /app/data /app/uploads

# Use non-root user for security
RUN adduser -D -u 1000 photouser && \
    chown -R photouser:photouser /app
USER photouser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./app"]
