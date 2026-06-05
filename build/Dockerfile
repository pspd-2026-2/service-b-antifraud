FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o antifraud-service ./cmd/server/main.go

FROM alpine:3.20
WORKDIR /root/

COPY --from=builder /app/antifraud-service .

COPY build/backup.json ./build/backup.json

EXPOSE 50052 8081
CMD ["./antifraud-service"]