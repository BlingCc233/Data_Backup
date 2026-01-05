# syntax=docker/dockerfile:1.7

# This image is primarily meant to BUILD the Linux binary of the Wails app.
# Running GUI apps inside Docker requires additional host setup (X11/Wayland forwarding).

ARG GO_VERSION=1.24
ARG WAILS_VERSION=v2.10.2
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

FROM --platform=$TARGETPLATFORM golang:${GO_VERSION}-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential pkg-config ca-certificates git \
    nodejs npm \
    libgtk-3-dev libwebkit2gtk-4.0-dev libayatana-appindicator3-dev \
  && rm -rf /var/lib/apt/lists/*

RUN go install github.com/wailsapp/wails/v2/cmd/wails@${WAILS_VERSION}

WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci

COPY . .

ENV CGO_ENABLED=1
RUN wails build -platform ${TARGETOS}/${TARGETARCH} -clean -ldflags="-w -s"

FROM --platform=$TARGETPLATFORM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libgtk-3-0 libwebkit2gtk-4.0-37 libayatana-appindicator3-1 \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/build/bin/go-backup-app /app/ccbackup

ENTRYPOINT ["/app/ccbackup"]
