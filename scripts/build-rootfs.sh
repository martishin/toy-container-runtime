#!/usr/bin/env bash
set -euo pipefail

IMAGE_TAG="${1:-mini/rootfs:alpine}"
ROOTFS_DIR="${2:-./rootfs}"

docker build -t "$IMAGE_TAG" .

CID="$(docker create "$IMAGE_TAG" sh -lc 'echo exporting rootfs')"

rm -rf "$ROOTFS_DIR"
mkdir -p "$ROOTFS_DIR"

TMP_TAR="$(mktemp -t rootfs.XXXXXX.tar)"
docker export "$CID" -o "$TMP_TAR"

# Important flags for Docker Desktop bind mounts on macOS:
#   --no-same-owner       -> don’t chown files to the archived UID/GID
#   --no-same-permissions -> don’t force exact mode bits (avoid chmod errors)
tar --no-same-owner --no-same-permissions -C "$ROOTFS_DIR" -xf "$TMP_TAR"

rm -f "$TMP_TAR"
docker rm "$CID" >/dev/null

echo "Rootfs exported to: $ROOTFS_DIR"
