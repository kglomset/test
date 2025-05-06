# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire application
COPY . .

# Build the application
# Targeting the main.go file in the cmd/server directory
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

# Final stage - using a minimal alpine image
FROM alpine:3.18

# Add necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Copy any config files if needed
# Uncomment and adjust if you have config files
# COPY --from=builder /app/config ./config

# Declare environment variables that will be provided by the host VM
ENV DB_HOST \
    DB_PORT \
    DB_USER \
    DB_PASSWORD \
    DB_NAME

# Expose the application port
# Adjust the port number if your application uses a different port
EXPOSE 8080

# Add a health check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:8080/health || exit 1

# Command to run the executable
CMD ["./server"]