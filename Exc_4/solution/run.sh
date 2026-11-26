#!/bin/sh

# todo
# docker build
# docker run db
# docker run orderservice
#!/bin/bash
set -e

# Compile the Go application for Linux architecture
CGO_ENABLED=0 GOOS=linux go build -o /app/ordersystem main.go
