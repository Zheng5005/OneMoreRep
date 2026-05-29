# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.3
ARG ALPINE_VERSION=3.21

# ------------------------------------------------------------------------------
# Base stage with common dependencies
# ------------------------------------------------------------------------------
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
RUN apk add --no-cache ca-certificates git make
WORKDIR /app

# ------------------------------------------------------------------------------
# Development stage with hot-reload (air)
# ------------------------------------------------------------------------------
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["air", "-c", ".air.toml"]

# ------------------------------------------------------------------------------
# Build stage
# ------------------------------------------------------------------------------
FROM base AS build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/api ./cmd/api

# ------------------------------------------------------------------------------
# Production stage
# ------------------------------------------------------------------------------
FROM scratch AS production
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
