FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git for dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Start a new stage from scratch
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy .env file if it exists
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]