# Stage 1: Build the Go application
FROM golang:1.22

# Set working directory
WORKDIR /app

# Install Goose migration tool
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy the schema folder into the image
COPY ./assets/schema /app/schema

# Set environment variables (can be overridden at runtime)
ENV DB_TYPE=""
ENV DB_CONNECTION_STRING=""
ENV MIGRATION_DIRECTION=""

# Entry point to run migrations
# $DB_TYPE : postgres
# $DB_CONNECTION_STRING : postgres://postgres:pass@localhost:5432/postgres
# $MIGRATION_DIRECTION : up (or) down
ENTRYPOINT ["/bin/bash", "-c", "goose -dir /app/schema $DB_TYPE $DB_CONNECTION_STRING $MIGRATION_DIRECTION"]
