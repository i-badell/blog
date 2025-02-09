# Build Stage
FROM golang:1.20-alpine AS builder

# Install git if needed (for go modules)
RUN apk add --no-cache git

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum and download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the Go application statically.
RUN CGO_ENABLED=0 GOOS=linux go build -a -o blogserver .

# Final Stage
FROM alpine:latest

# Install certificates if your app makes outbound TLS calls (optional)
RUN apk --no-cache add ca-certificates

# Set the working directory in the final container.
WORKDIR /root/

# Copy the built binary and required folders from the builder stage.
COPY --from=builder /app/blogserver .
COPY --from=builder /app/public ./public
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/posts ./posts

# Expose the port your app is listening on.
EXPOSE 8080

# Start the application.
CMD ["./blogserver"]
