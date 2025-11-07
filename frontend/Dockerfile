# Multi-stage build for optimized production image

# Stage 1: Builder - Build the React application
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with clean install for reproducible builds
RUN npm ci --silent

# Copy application source code
COPY . .

# Build application for production
RUN npm run build

# Stage 2: Runtime - Serve with Nginx
FROM nginx:alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Start Nginx server
CMD ["nginx", "-g", "daemon off;"]

# Labels for metadata
LABEL maintainer="pedidosonline@example.com"
LABEL description="Frontend React application for Pedidos Online"
LABEL version="1.0.0"
