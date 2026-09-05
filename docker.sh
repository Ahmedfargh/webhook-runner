#!/usr/bin/env bash

# ==============================================================================
# Webhook Microservices Docker Management CLI
# ==============================================================================

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

ACTION=${1:-help}
SERVICE=$2

print_banner() {
    echo -e "${BOLD}${CYAN}===================================================================${NC}"
    echo -e "${BOLD}${CYAN}   🐳 Webhook Microservices Ecosystem Docker CLI                  ${NC}"
    echo -e "${BOLD}${CYAN}===================================================================${NC}"
}

case "$ACTION" in
    build)
        print_banner
        echo -e "${BLUE}🔨 Building all lightweight Docker container images...${NC}"
        DOCKER_BUILDKIT=1 docker compose build
        echo -e "${GREEN}✓ All images built successfully!${NC}"
        ;;

    up|start)
        print_banner
        echo -e "${PURPLE}🚀 Starting Webhook microservices ecosystem in background...${NC}"
        docker compose up -d
        echo -e "${GREEN}✓ Containers deployed. Checking status...${NC}"
        sleep 2
        docker compose ps
        echo -e "\n${BOLD}${GREEN}🌐 Access Points:${NC}"
        echo -e "  Frontend UI:       ${BOLD}http://localhost:5173${NC}"
        echo -e "  API Gateway REST:  ${BOLD}http://localhost:8080${NC}"
        echo -e "  Gateway Health:    ${BOLD}http://localhost:8080/health${NC}"
        echo -e "  Accounts gRPC:     ${BOLD}localhost:50051${NC}"
        echo -e "  Subscriptions gRPC:${BOLD}localhost:50052${NC}"
        echo -e "  MySQL Port:        ${BOLD}localhost:3307${NC}"
        ;;

    down|stop)
        print_banner
        echo -e "${YELLOW}🛑 Stopping all containers gracefully...${NC}"
        docker compose down
        echo -e "${GREEN}✓ All containers stopped.${NC}"
        ;;

    restart)
        print_banner
        echo -e "${YELLOW}🔄 Restarting containers...${NC}"
        docker compose restart $SERVICE
        echo -e "${GREEN}✓ Restart complete.${NC}"
        ;;

    logs)
        if [ -n "$SERVICE" ]; then
            docker compose logs -f "$SERVICE"
        else
            docker compose logs -f
        fi
        ;;

    ps|status)
        print_banner
        docker compose ps
        ;;

    clean)
        print_banner
        echo -e "${RED}⚠️  Cleaning containers, images and volumes...${NC}"
        docker compose down -v --rmi local
        echo -e "${GREEN}✓ Docker environment cleaned.${NC}"
        ;;

    help|*)
        print_banner
        echo -e "Usage: ${BOLD}./docker.sh [command]${NC}\n"
        echo -e "Available commands:"
        echo -e "  ${CYAN}build${NC}             Build lightweight multi-stage Docker images"
        echo -e "  ${CYAN}up / start${NC}        Start all microservices in the background"
        echo -e "  ${CYAN}down / stop${NC}       Stop all running containers"
        echo -e "  ${CYAN}restart [svc]${NC}     Restart all or a specific service (e.g., ./docker.sh restart api-gateway)"
        echo -e "  ${CYAN}logs [svc]${NC}        Follow log output (e.g., ./docker.sh logs accounts)"
        echo -e "  ${CYAN}status / ps${NC}       List running containers and their health"
        echo -e "  ${CYAN}clean${NC}             Tear down containers, local images, and volumes"
        echo ""
        ;;
esac
