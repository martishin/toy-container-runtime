# toy-container-runtime 🚗

A tiny container runtime written in Go that demonstrates **how containers actually work** - no magic, just Linux
features:

- **Namespaces** (UTS, PID, Mount) → isolate what a process can *see*
- **chroot + rootfs** → give the process its own filesystem view
- **cgroups (pids)** → limit what a process can *use*
- Minimal **mount setup** → `/proc` and a scratch `tmpfs`


## Why containers?

Containers are *just Linux processes* with:

- a **restricted view** (namespaces),
- a **different root filesystem** (chroot/pivot_root with a “rootfs”),
- and **resource limits** (cgroups).

This makes them **portable** (ship a filesystem), **isolated** (can’t see host stuff), and **efficient** (no full VM per
app). Real runtimes (runc/containerd/Docker) add a ton of safety and orchestration around these same primitives.


## What this project demonstrates

1) **UTS namespace** → custom hostname (`container`).
2) **PID namespace** → your process becomes **PID 1** inside the container.
3) **Mount namespace** → private mount table (no host pollution).
4) **chroot rootfs** → your command runs in a Docker-built filesystem.
5) **/proc + tmpfs** → mounts that userspace tools rely on.
6) **pids cgroup** → best-effort process count limit.

You’ll literally see: a different hostname, PID 1, your own `/proc`, and a private tmpfs.


## How it works

- We build a root filesystem from `Dockerfile` (Alpine + python3 + bash), then `docker export` it to `./rootfs`.
- `mini_container run ...` re-execs itself as `child` with:
    - `CLONE_NEWUTS | CLONE_NEWPID | CLONE_NEWNS` (+ `Unshare(CLONE_NEWNS)`)
    - set hostname → `container`
    - `chroot(rootfs)` and `chdir("/")`
    - make `/` a mountpoint, set mount propagation (slave → private)
    - mount `/proc` and a `tmpfs` at `/mytemp`
    - (best-effort) put the process in a **pids cgroup** to cap forks
- Finally, it `exec`s your command (`/usr/bin/python3 /opt/app/hello.py` or `/bin/bash`).


## Prerequisites

- Linux (or macOS via Dev Containers / Docker Desktop)
- Go 1.24.5
- Docker (to build and export the rootfs)
- In devcontainers: container needs `--privileged`, `SYS_ADMIN`, and the Docker socket mounted


## Running locally

```bash
# Build the runtime
make build

# One-shot: build image -> export rootfs -> run demo app
make run
# Expected: "Hello from container!"

# Interactive shell inside the container
make bash
```


## Verify you are in a container

Inside the container shell:

```bash
hostname                    # -> container
mount | head                # shows /proc and /mytemp tmpfs
ps -ef                      # PID namespace view (you're PID 1)
cat /proc/self/status | sed -n '1,12p'
```


## Project structure

```
.
├── Dockerfile                   # rootfs content (Alpine + python3 + bash)
├── Makefile
├── app/hello.py                 # demo app
├── cmd/mini_container/main.go   # Cobra CLI: build/run/run-rootfs/child
├── internal/
│   ├── build/dockerfile_driver.go  # docker build/create/export -> ./rootfs
│   └── container/
│       ├── namespaces.go           # SysProcAttr (linux)
│       ├── fs.go                   # mount propagation, /proc, tmpfs
│       ├── cgroups.go              # pids cgroup (v1/v2 best-effort)
│       └── runner.go               # Run / RunChild
├── scripts/build-rootfs.sh
└── go.mod / go.sum
```
