# ==============================================================================
# API Gateway Microservice - Lightweight Multi-Stage Dockerfile
# ==============================================================================

# Build Stage
FROM golang:alpine AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache ca-certificates git tzdata

# Cache Go modules layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile statically linked, stripped lightweight binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/server cmd/server/main.go

# Production Stage (Minimal Alpine Runtime ~7MB)
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata curl \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server /app/server

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]
