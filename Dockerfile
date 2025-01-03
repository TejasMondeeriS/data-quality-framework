# Stage 1: Build the Go application
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN cd cmd/api && go build -o /app/bin/api

# Stage 2: Create a minimal image for running the Go application
FROM ubuntu:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/api /app/api

# Expose the port the application will run on
EXPOSE 4444

# Command to run the application
ENTRYPOINT ["/app/api"]