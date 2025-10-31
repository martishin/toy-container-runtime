BIN ?= ./bin/mini_container
IMAGE_TAG ?= mini/rootfs:alpine
ROOTFS_DIR ?= ./rootfs

.PHONY: build
build:
	go build -o $(BIN) ./cmd/mini_container

.PHONY: rootfs
rootfs:
	./scripts/build-rootfs.sh $(IMAGE_TAG)

.PHONY: run
run: build rootfs
	sudo $(BIN) run-rootfs $(ROOTFS_DIR) -- /usr/bin/python3 /opt/app/hello.py

.PHONY: bash
bash: build rootfs
	sudo $(BIN) run-rootfs $(ROOTFS_DIR) -- /bin/bash
