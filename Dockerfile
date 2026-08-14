### Stage 1: build frontend
FROM oven/bun:1.3.14 AS frontend
WORKDIR /app/frontend
ENV NODE_OPTIONS=--no-deprecation
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY frontend/ ./
RUN bun run build

### Stage 2: build Go binary
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
ARG VERSION=dev
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN rm -rf internal/webui/dist/*
COPY --from=frontend /app/frontend/dist ./internal/webui/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o booksync .

### Stage 3: runtime
FROM alpine:3.24
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/booksync .

ENV BOOKSYNC_DATA_DIR=/config

EXPOSE 8686
VOLUME ["/config"]

ENTRYPOINT ["/app/booksync"]
CMD ["serve"]
