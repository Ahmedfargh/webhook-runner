#!/usr/bin/env bash

# ==============================================================================
# Webhook Microservices Ecosystem Runner
# Starts: Accounts, Subscriptions, Webhook Runner, Audit Service, API Gateway, Frontend
# ==============================================================================

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
MAGENTA='\033[0;95m'
NC='\033[0m' # No Color
BOLD='\033[1m'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BOLD}${CYAN}===================================================================${NC}"
echo -e "${BOLD}${CYAN}  🚀 Starting Webhook Microservices Ecosystem (Zoho UI & Gateway)  ${NC}"
echo -e "${BOLD}${CYAN}===================================================================${NC}"

# Ensure MySQL databases exist
if command -v mysql &> /dev/null; then
    mysql -u root -e "CREATE DATABASE IF NOT EXISTS webhook_accounts; CREATE DATABASE IF NOT EXISTS webhook_subscriptions; CREATE DATABASE IF NOT EXISTS webhook_runner; CREATE DATABASE IF NOT EXISTS webhook_audit;" 2>/dev/null || true
elif command -v mariadb &> /dev/null; then
    mariadb -u root -e "CREATE DATABASE IF NOT EXISTS webhook_accounts; CREATE DATABASE IF NOT EXISTS webhook_subscriptions; CREATE DATABASE IF NOT EXISTS webhook_runner; CREATE DATABASE IF NOT EXISTS webhook_audit;" 2>/dev/null || true
fi

# Ensure Kafka broker is running or prompt user
KAFKA_RUNNING=false
if (echo > /dev/tcp/localhost/9092) 2>/dev/null; then
    KAFKA_RUNNING=true
elif command -v nc &> /dev/null && nc -z -w 1 localhost 9092 2>/dev/null; then
    KAFKA_RUNNING=true
fi

if [ "$KAFKA_RUNNING" = false ]; then
    if command -v docker &> /dev/null; then
        echo -e "${YELLOW}[Kafka] Broker not detected on localhost:9092. Attempting to start Kafka container...${NC}"
        (docker compose up -d kafka 2>/dev/null || docker-compose up -d kafka 2>/dev/null || true)
        sleep 2
    else
        echo -e "${YELLOW}[Kafka Notice] Kafka broker is not running on localhost:9092.${NC}"
        echo -e "${YELLOW}  👉 Run 'docker compose up -d kafka' to start Kafka broker, or set KAFKA_ENABLED=false in .env${NC}"
    fi
fi

# Ensure .env files exist
if [ ! -f "$ROOT_DIR/accounts/.env" ]; then
    echo -e "${YELLOW}[Notice] accounts/.env not found, generating default config...${NC}"
    cat <<EOF > "$ROOT_DIR/accounts/.env"
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=webhook_accounts
GRPC_PORT=50051
AUTH_TOKEN=4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1
ALLOWED_SERVICES=api-gateway,webhook-runner,audit-service
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_AUDIT_EVENTS=audit-events
KAFKA_ENABLED=true
EOF
fi

if [ ! -f "$ROOT_DIR/subscriptions/.env" ]; then
    echo -e "${YELLOW}[Notice] subscriptions/.env not found, generating default config...${NC}"
    cat <<EOF > "$ROOT_DIR/subscriptions/.env"
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=webhook_subscriptions
GRPC_PORT=50052
AUTH_TOKEN=4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1
ALLOWED_SERVICES=api-gateway,webhook-runner,audit-service
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_AUDIT_EVENTS=audit-events
KAFKA_ENABLED=true
EOF
fi

if [ ! -f "$ROOT_DIR/webhook-runner/.env" ]; then
    echo -e "${YELLOW}[Notice] webhook-runner/.env not found, generating default config...${NC}"
    cat <<EOF > "$ROOT_DIR/webhook-runner/.env"
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=webhook_runner
GRPC_PORT=50053
AUTH_TOKEN=4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1
ALLOWED_SERVICES=api-gateway,webhook-runner,audit-service
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_WEBHOOK_DISPATCH=webhook-dispatches
KAFKA_TOPIC_WEBHOOK_RESULTS=webhook-results
KAFKA_TOPIC_AUDIT_EVENTS=audit-events
KAFKA_GROUP_ID=webhook-runner-group
KAFKA_ENABLED=true
EOF
fi

if [ ! -f "$ROOT_DIR/audit-service/.env" ]; then
    echo -e "${YELLOW}[Notice] audit-service/.env not found, generating default config...${NC}"
    cat <<EOF > "$ROOT_DIR/audit-service/.env"
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=webhook_audit
GRPC_PORT=50054
AUTH_TOKEN=4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1
ALLOWED_SERVICES=api-gateway,accounts,subscriptions,webhook-runner,audit-service
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_AUDIT_EVENTS=audit-events
KAFKA_GROUP_ID=audit-service-group
KAFKA_ENABLED=true
EOF
fi

if [ ! -f "$ROOT_DIR/api-gateway/.env" ]; then
    echo -e "${YELLOW}[Notice] api-gateway/.env not found, generating default config...${NC}"
    cat <<EOF > "$ROOT_DIR/api-gateway/.env"
PORT=8080
ACCOUNTS_GRPC_HOST=localhost
ACCOUNTS_GRPC_PORT=50051
SUBSCRIPTIONS_GRPC_HOST=localhost
SUBSCRIPTIONS_GRPC_PORT=50052
RUNNER_GRPC_HOST=localhost
RUNNER_GRPC_PORT=50053
AUDIT_GRPC_HOST=localhost
AUDIT_GRPC_PORT=50054
SERVICE_NAME=api-gateway
SERVICE_TOKEN=4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1
JWT_SECRET=api-gateway-super-secret-jwt-key-2026
ALLOWED_ORIGINS=*
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_WEBHOOK_DISPATCH=webhook-dispatches
KAFKA_TOPIC_WEBHOOK_RESULTS=webhook-results
KAFKA_TOPIC_AUDIT_EVENTS=audit-events
KAFKA_ENABLED=true
EOF
fi

if [ ! -f "$ROOT_DIR/frontend/.env" ]; then
    echo -e "${YELLOW}[Notice] frontend/.env not found, generating default config...${NC}"
    cat <<EOF > "$ROOT_DIR/frontend/.env"
VITE_API_URL=http://localhost:8080/api/v1
EOF
fi

# Track child PIDs
PIDS=()

cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down all microservices gracefully...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null || true
        fi
    done
    wait 2>/dev/null || true
    echo -e "${GREEN}✓ All services stopped.${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# 1. Start Accounts gRPC Service
echo -e "${BLUE}▶ [1/6] Starting Accounts gRPC Service on port 50051...${NC}"
(
    cd "$ROOT_DIR/accounts"
    go run cmd/server/main.go 2>&1 | sed -e "s/^/$(printf "${BLUE}[Accounts gRPC]${NC}      ")/"
) &
PIDS+=($!)

# 2. Start Subscriptions gRPC Service
echo -e "${PURPLE}▶ [2/6] Starting Subscriptions gRPC Service on port 50052...${NC}"
(
    cd "$ROOT_DIR/subscriptions"
    go run cmd/server/main.go 2>&1 | sed -e "s/^/$(printf "${PURPLE}[Subscriptions gRPC]${NC} ")/"
) &
PIDS+=($!)

# 3. Start Webhook Runner gRPC Service
echo -e "${YELLOW}▶ [3/6] Starting Webhook Runner gRPC Service on port 50053...${NC}"
(
    cd "$ROOT_DIR/webhook-runner"
    go run cmd/server/main.go 2>&1 | sed -e "s/^/$(printf "${YELLOW}[Runner gRPC]${NC}        ")/"
) &
PIDS+=($!)

# 4. Start Audit Microservice & Kafka Worker
echo -e "${MAGENTA}▶ [4/6] Starting Audit Microservice & Kafka Worker on port 50054...${NC}"
(
    cd "$ROOT_DIR/audit-service"
    go run cmd/server/main.go 2>&1 | sed -e "s/^/$(printf "${MAGENTA}[Audit Service]${NC}      ")/"
) &
PIDS+=($!)

# Give gRPC servers a moment to bind ports
sleep 1.5

# 5. Start API Gateway
echo -e "${GREEN}▶ [5/6] Starting API Gateway HTTP REST on port 8080...${NC}"
(
    cd "$ROOT_DIR/api-gateway"
    go run cmd/server/main.go 2>&1 | sed -e "s/^/$(printf "${GREEN}[API Gateway]${NC}        ")/ "
) &
PIDS+=($!)

# Give Gateway a moment to bind port
sleep 1

# 6. Start Frontend Dev Server
echo -e "${CYAN}▶ [6/6] Starting Vue 3 Frontend (Zoho Projects UI) on port 5173...${NC}"
(
    cd "$ROOT_DIR/frontend"
    npm run dev -- --host 2>&1 | sed -e "s/^/$(printf "${CYAN}[Frontend UI]${NC}        ")/"
) &
PIDS+=($!)

echo -e "\n${BOLD}${GREEN}===================================================================${NC}"
echo -e "${BOLD}${GREEN}  ✨ All Microservices Running Successfully!                       ${NC}"
echo -e "${BOLD}${GREEN}===================================================================${NC}"
echo -e "  🌐 ${BOLD}Frontend UI:${NC}         http://localhost:5173"
echo -e "  🚪 ${BOLD}API Gateway REST:${NC}    http://localhost:8080"
echo -e "  🩺 ${BOLD}Gateway Health:${NC}      http://localhost:8080/health"
echo -e "  🔒 ${BOLD}Accounts gRPC:${NC}       localhost:50051 (Service Auth Enforced)"
echo -e "  💳 ${BOLD}Subscriptions gRPC:${NC}  localhost:50052 (Service Auth Enforced)"
echo -e "  ⚡ ${BOLD}Webhook Runner gRPC:${NC} localhost:50053 (Service Auth Enforced)"
echo -e "  📜 ${BOLD}Audit Service gRPC:${NC}  localhost:50054 (Service Auth Enforced)"
echo -e "  📬 ${BOLD}Kafka Event Stream:${NC}  localhost:9092 (Topics: webhook-dispatches, audit-events)"
echo -e "${BOLD}${GREEN}===================================================================${NC}"
echo -e "Press ${BOLD}Ctrl+C${NC} to stop all services.\n"

# Wait for background processes
wait
