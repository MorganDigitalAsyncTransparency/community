# Go backend — multi-stage build
#
# Stage 1: compile the Go binary in golang:alpine
# Stage 2: copy the binary into a minimal Alpine runtime image
#
# The backend polls the Discourse API, runs observation logic,
# and serves the internal API on port 8080.
# It is not exposed externally — nginx proxies /api/ requests to it.
#
# CGO_ENABLED will need to be 1 once SQLite is added (ADR 0006).
# Until then, pure-Go build with CGO disabled.

# -- build stage --
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN [ -f go.sum ] && go mod download || true
COPY src/ src/
RUN CGO_ENABLED=0 go build -o /app/discourse-observer ./src/...

# -- runtime stage --
FROM alpine:3
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/discourse-observer .
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./discourse-observer"]
