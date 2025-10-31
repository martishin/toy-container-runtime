FROM golang:1.24.5-bullseye

# Install build tools, Docker CLI (used by your rootfs builder), and Python3
RUN apt-get update && apt-get install -y --no-install-recommends \
      make git curl python3 docker.io \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work
