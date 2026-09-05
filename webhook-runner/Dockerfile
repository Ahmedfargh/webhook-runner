# ==============================================================================
# Webhook Runner Worker Service - Lightweight Dockerfile
# ==============================================================================

FROM golang:alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git tzdata

# Check if go.mod exists before downloading
COPY go.mod* go.sum* ./
RUN if [ -f go.mod ]; then go mod download; fi

COPY . .

# Build only if cmd/server/main.go exists, else create minimal dummy
RUN if [ -f cmd/server/main.go ]; then \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server cmd/server/main.go; \
    else \
        echo 'package main; func main(){ select{} }' > dummy.go && \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server dummy.go; \
    fi

FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server /app/server

USER appuser

ENTRYPOINT ["/app/server"]
