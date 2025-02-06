FROM golang:1.23.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/scanner ./scanner/main.go

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 appuser
USER appuser

COPY --from=builder --chown=appuser:appuser /app/bin/scanner /app/scanner
COPY --chown=appuser:appuser config /app/config

WORKDIR /app

ENTRYPOINT ["/app/scanner"]
