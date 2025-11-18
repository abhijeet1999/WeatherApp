# Multi-stage build for single container with both services
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY Producer/ ./Producer/
COPY Consumer/ ./Consumer/
COPY models/ ./models/

# Build both services with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -trimpath -o producer ./Producer/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -trimpath -o consumer ./Consumer/main.go

# Final stage - Optimized Alpine
FROM alpine:3.19

# Install only essential packages
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy built binaries
COPY --from=builder /app/producer .
COPY --from=builder /app/consumer .

# Copy input file and create optimized startup script
COPY input.txt .
COPY <<EOF start.sh
#!/bin/sh
echo "🚀 Starting WeatherApp Services..."
echo "📤 Starting Producer..."
./producer &
sleep 5
echo "📥 Starting Consumer..."
./consumer &
echo "✅ Both services started!"
echo "⏹️  Press Ctrl+C to stop..."
wait
EOF

# Set permissions and ownership
RUN chmod +x start.sh && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 8081

# Start both services
CMD ["./start.sh"]

