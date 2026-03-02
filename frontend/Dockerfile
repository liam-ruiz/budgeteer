# Stage 1: Build the Angular app
FROM node:22-alpine AS builder

WORKDIR /app

# Copy dependency files first for better caching
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and build for production
COPY . .
RUN npx ng build --configuration production

# Stage 2: Serve with nginx
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder stage
COPY --from=builder /app/dist/budget/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
