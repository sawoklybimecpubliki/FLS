# Build stage using Go 1.23
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Add and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the full source
COPY . .

# Build the binary (change path if needed)
RUN go build -o us ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /

# Copy config
COPY --from=builder /app/cmd/api/config.yaml .

# Copy binary
COPY --from=builder /app/us .

EXPOSE 8080

CMD ["./us"]