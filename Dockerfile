# Use the official Golang image as a build stage
FROM golang:1.23-alpine3.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy the go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

# Debug: List files in the /app directory
RUN ls -l /app

# Use a minimal base image for the final stage
FROM alpine:latest

# Set the working directory for the final image
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]