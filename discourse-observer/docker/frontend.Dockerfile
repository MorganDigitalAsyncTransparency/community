# React frontend — multi-stage build
#
# Stage 1: build the React app with Node
# Stage 2: serve the static output with nginx
#
# Nginx serves the React build on port 80 and proxies /api/ to the backend.
# Exposed on host port 3000 via docker-compose.

# -- build stage --
FROM node:22-alpine AS build
WORKDIR /src
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ .
RUN npm run build

# -- runtime stage --
FROM nginx:alpine
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80
