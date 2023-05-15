# Stage 1: Build stage
FROM golang:1.20 AS build

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o server

# Stage 2: Final stage
FROM alpine:3.18

# Install additional tools for debugging
RUN apk --no-cache add bash curl jq

WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build /app/server .

# Expose the server port
EXPOSE 8080

# Run the server binary
CMD ["./server"]
