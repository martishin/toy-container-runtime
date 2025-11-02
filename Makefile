BIN        ?= ./bin/mini_container
IMAGE_TAG  ?= mini/rootfs:alpine
ROOTFS_DIR ?= ./rootfs

.PHONY: build
build:
	GOOS=linux go build -o $(BIN) ./cmd/mini_container

.PHONY: rootfs
rootfs:
	bash ./scripts/build-rootfs.sh $(IMAGE_TAG) $(ROOTFS_DIR)

# Runs using an already-exported ./rootfs
.PHONY: run-rootfs
run-rootfs: build rootfs
	$(BIN) run-rootfs $(ROOTFS_DIR) -- /usr/bin/python3 /opt/app/hello.py

# One-shot: build image -> export rootfs -> run (no separate rootfs step needed)
.PHONY: run
run: build
	$(BIN) run -f Dockerfile -C . -- /usr/bin/python3 /opt/app/hello.py

.PHONY: bash
bash: build rootfs
	$(BIN) run-rootfs $(ROOTFS_DIR) -- /bin/bash
