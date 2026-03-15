# React frontend — multi-stage build
#
# Stage 1: build the React app with Node
# Stage 2: serve the static output with nginx
#
# Nginx serves the React build on port 80 and proxies /api/ to the backend.
# Exposed on host port 3000 via docker-compose.

# -- build stage --
FROM node:24-alpine AS build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# -- runtime stage --
FROM nginx:alpine
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80
