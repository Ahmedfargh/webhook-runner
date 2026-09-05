# ==============================================================================
# Vue 3 Frontend - Lightweight Multi-Stage Nginx Dockerfile (~25MB)
# ==============================================================================

# Build Stage
FROM node:22-alpine AS builder

WORKDIR /app

# Cache node modules layer
COPY package.json package-lock.json* ./
RUN npm ci || npm install

# Pass build-time environment variable for API URL (defaults to /api/v1 for reverse proxy)
ARG VITE_API_URL=http://localhost:8080/api/v1
ENV VITE_API_URL=$VITE_API_URL

COPY . .
RUN npm run build

# Production Stage (Ultra-light Alpine Nginx)
FROM nginx:alpine

# Remove default nginx webpage
RUN rm -rf /usr/share/nginx/html/*

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy production build assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O - http://localhost/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]
