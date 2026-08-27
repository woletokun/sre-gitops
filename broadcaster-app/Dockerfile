FROM golang:1.22-alpine AS builder
WORKDIR /app

# Copy module definition files
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o broadcast-server cmd/main.go

# Production runtime image
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/broadcast-server /usr/local/bin/broadcast-server

EXPOSE 8080 8081
ENTRYPOINT ["broadcast-server"]
CMD ["start", "8080"]
