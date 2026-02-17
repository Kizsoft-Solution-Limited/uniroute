# Frontend - build Vite app and serve with nginx
# Build context must be repo root: docker build -f frontend/Dockerfile .
FROM node:20-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
ARG VITE_API_BASE_URL=https://app.uniroute.co
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
# Use vite build only in Docker to avoid vue-tsc/TypeScript compatibility failure
RUN npx vite build

FROM nginx:alpine
RUN rm -rf /usr/share/nginx/html/*
COPY --from=builder /app/dist /usr/share/nginx/html
COPY --from=builder /app/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
