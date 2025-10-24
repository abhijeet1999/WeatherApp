# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY Producer/ ./Producer/
COPY models/ ./models/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o producer ./Producer/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/producer .

# Copy input file
COPY input.txt .

# Expose port (if needed for health checks)
EXPOSE 8080

# Command to run
CMD ["./producer"]
