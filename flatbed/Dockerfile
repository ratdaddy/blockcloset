# syntax=docker/dockerfile:1.7

### 1) Build stage
FROM golang:1.24-bookworm AS build
WORKDIR /src

# Build metadata (optional, nice for --version endpoints/logs)
ARG VERSION=dev
ARG COMMIT=dirty
ARG DATE=unknown

# Module deps first (better cache)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# App source
COPY . .

# Static binary, small & portable
ENV CGO_ENABLED=0
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
      -o /out/server ./cmd/flatbed

### 2) Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
