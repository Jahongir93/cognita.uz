# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gogame ./cmd/server/main.go

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/gogame .
RUN mkdir -p uploads

EXPOSE 8080
ENTRYPOINT ["./gogame"]
