# Builder stage
FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o cloud ./main/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app .

EXPOSE 80

ENTRYPOINT ["/app/cloud"]
