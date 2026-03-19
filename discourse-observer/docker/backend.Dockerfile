# Go backend — multi-stage build
#
# Stage 1: compile the Go binary in golang:alpine
# Stage 2: copy the binary into a minimal Alpine runtime image
#
# The backend polls the Discourse API, runs observation logic,
# and serves the internal API on port 8080.
# It is not exposed externally — nginx proxies /api/ requests to it.
#
# SQLite is provided by modernc.org/sqlite (pure Go, no CGO needed).
# CGO_ENABLED=0 produces a static binary.

# -- build stage --
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN [ -f go.sum ] && go mod download || true
COPY backend/ backend/
COPY config/ config/
RUN CGO_ENABLED=0 go build -o /app/discourse-observer ./backend

# -- runtime stage --
FROM alpine:3.19

WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/discourse-observer .
COPY config/ config/
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./discourse-observer"]
