# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Copy the vendor directory which contains all dependencies
COPY vendor ./vendor

# (Optional but good practice) Verify that the dependencies in vendor match go.mod
RUN go mod verify

# Copy the rest of the application source code
COPY . .

# Build the application, telling Go to use the vendor directory.
# The -mod=vendor flag is CRITICAL.
# The path ./cmd/app/main.go IS CORRECT because that's your entry point.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/app ./cmd/app/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/app .

# Copy environment example file (for reference if needed inside container)
COPY .env.example .env.example

# Expose the port your application will run on
EXPOSE 8080

# Command to run the executable
CMD ["./app"]