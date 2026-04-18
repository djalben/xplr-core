# ===================================
# Stage 1: Build the React application
# ===================================
FROM node:22-alpine AS builder

# Build ID for cache busting
ARG BUILD_ID=unknown
ENV BUILD_ID=${BUILD_ID}

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (including devDependencies for build)
RUN npm ci

# Copy source code
COPY . .

# CRITICAL: Remove any existing dist folder before build
RUN rm -rf dist build web-build .expo

# Build the application for production
RUN npm run build

# Verify build succeeded
RUN test -f dist/index.html || (echo "ERROR: Build failed - dist/index.html not found" && exit 1)

# Fix script tags to support ES modules
RUN sed -i 's/<script src=/<script type="module" src=/g' dist/index.html

# ===================================
# Stage 2: Serve with NGINX
# ===================================
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/

# Copy built files from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
