# Go backend — multi-stage build
#
# Stage 1: compile Go binary
# Stage 2: copy binary into minimal Alpine image
#
# The backend polls the Discourse API, runs observation logic,
# and serves the internal API on port 8080.
# It is not exposed externally — nginx proxies /api/ requests to it.

# -- build stage --
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN [ -f go.sum ] && go mod download || true
COPY src/ src/
RUN CGO_ENABLED=1 go build -o /app/discourse-observer ./src/...

# -- runtime stage --
FROM alpine:3
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/discourse-observer .
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./discourse-observer"]
