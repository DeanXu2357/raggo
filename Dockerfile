# Start from the latest golang base image
FROM golang:1.23.4-alpine3.21 AS build-env

# Add Maintainer Info
LABEL maintainer="DeanXu2357 <dean.xu.2357@gmail.com>"

RUN apk --no-cache add curl

# Install delve
# Warning: This is not necessary for production. It is only for debugging purposes.
RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM build-env

# Set the Current Working Directory inside the Docker container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the Docker container
COPY . .

# Copy .env.yaml from the current directory to /root directory and rename it to .dispatchapi.yaml inside the Docker container
RUN cp .env.yaml $HOME/.raggo.yaml

# Command to run the executable
# Warngin: In production, you should build the binary and run it instead of using go run.
CMD ["go", "run", "main.go", "serve"]
