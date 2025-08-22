# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app

# Enable modules and cache deps
COPY go.mod .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download -x || true

# Copy source
COPY . .

# Build static binary for linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

# Runtime stage (distroless)
FROM gcr.io/distroless/static-debian12:nonroot

ENV PORT=8080
EXPOSE 8080

COPY --from=build /out/server /server

USER nonroot:nonroot
ENTRYPOINT ["/server"]
