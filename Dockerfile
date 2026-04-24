# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first to leverage layer caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/panini-api \
    ./cmd/api/

# ---- Runtime stage ----
FROM alpine:3.20

# Install CA certificates for outbound TLS (e.g. APNs).
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user.
RUN addgroup -S panini && adduser -S panini -G panini

WORKDIR /app

COPY --from=builder /app/panini-api .

USER panini

EXPOSE 9090

ENTRYPOINT ["/app/panini-api"]
