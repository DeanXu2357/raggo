# Build stage
FROM golang:1.23.4-alpine3.21 AS builder

# Add Maintainer Info
LABEL maintainer="DeanXu2357 <dean.xu.2357@gmail.com>"

# Install build dependencies
RUN apk --no-cache add curl

# Install delve
# Warning: This is not necessary for production. It is only for debugging purposes.
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Command to run the executable
# Warning: In production, you should build the binary and run it instead of using go run.
CMD ["go", "run", "main.go", "serve"]
