# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git tzdata ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api/main.go

# Production Stage
FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/api /app/api

EXPOSE 8081

CMD ["/app/api"]
