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

# Build both services
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o producer ./Producer/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer ./Consumer/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /root/

# Copy built binaries
COPY --from=builder /app/producer .
COPY --from=builder /app/consumer .

# Copy input file
COPY input.txt .

# Create startup script
RUN echo '#!/bin/sh' > start.sh && \
    echo 'echo "🚀 Starting WeatherApp Services..."' >> start.sh && \
    echo 'echo "📤 Starting Producer..."' >> start.sh && \
    echo './producer &' >> start.sh && \
    echo 'sleep 5' >> start.sh && \
    echo 'echo "📥 Starting Consumer..."' >> start.sh && \
    echo './consumer &' >> start.sh && \
    echo 'echo "✅ Both services started!"' >> start.sh && \
    echo 'echo "⏹️  Press Ctrl+C to stop..."' >> start.sh && \
    echo 'wait' >> start.sh && \
    chmod +x start.sh

# Expose ports
EXPOSE 8080 8081

# Start both services
CMD ["./start.sh"]

