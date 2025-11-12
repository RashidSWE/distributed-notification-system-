
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make ca-certificates

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/push-service ./cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl bash tzdata

WORKDIR /app

COPY --from=builder /app/push-service .


COPY .env /app/.env


COPY firebase-credentials.json /app/firebase-credentials.json

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/api/health || exit 1

CMD ["/app/push-service"]
