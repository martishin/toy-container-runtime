#!/usr/bin/env bash
set -euo pipefail

IMAGE_TAG="${1:-mini/rootfs:alpine}"
ROOTFS_DIR="${2:-./rootfs}"

docker build -t "$IMAGE_TAG" .
CID="$(docker create "$IMAGE_TAG" sh -lc 'echo exporting rootfs')"

mkdir -p "ROOTFS_DIR"
docker export "$CID" | tar -C "$ROOTFS_DIR" -xf -
docker rm "$CID" > /dev/null

echo "Rootfs exported to: $ROOTFS_DIR"
