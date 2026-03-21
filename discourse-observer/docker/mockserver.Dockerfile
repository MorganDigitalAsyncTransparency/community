# Mock Discourse server — multi-stage build
#
# Spec: specs/observer/mock-server-service.md (MS-4)
#
# Serves realistic Discourse API responses from built-in fixtures.
# Used in dev mode so the sync pipeline works without a real forum.
# Listens on port 9920.

# -- build stage --
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN [ -f go.sum ] && go mod download || true
COPY backend/ backend/
RUN CGO_ENABLED=0 go build -o /app/mockserver ./backend/cmd/mockserver

# -- runtime stage --
FROM alpine:3.19

WORKDIR /app
COPY --from=build /app/mockserver .
EXPOSE 9920
CMD ["./mockserver"]
