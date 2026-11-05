  # todo
# --- Stage 1: Builder ---
FROM golang:1.27 AS builder

WORKDIR /app

# Copy all project files into the builder container
COPY . .

# Run the provided build script
RUN chmod +x scripts/build-application.sh && ./scripts/build-application.sh

# --- Stage 2: Runtime ---
FROM alpine:latest

WORKDIR /root/

# Install CA certificates in case your app makes external requests
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /app/ordersystem /root/ordersystem

# Expose the application port if needed (adjust if your app uses a specific port)
EXPOSE 8080

# Command to run the binary
ENTRYPOINT ["./ordersystem"]