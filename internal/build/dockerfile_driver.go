package build

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Options struct {
	ContextDir   string
	Dockerfile   string
	ImageTag     string
	OutputRootfs string
}

func BuildRootfsWithDocker(opts Options) error {
	if opts.ContextDir == "" {
		opts.ContextDir = "."
	}
	if opts.Dockerfile == "" {
		opts.Dockerfile = "Dockerfile"
	}
	if opts.ImageTag == "" {
		opts.ImageTag = "mini/rootfs:latest"
	}
	if opts.OutputRootfs == "" {
		opts.OutputRootfs = "./rootfs"
	}

	// 1) docker build
	build := exec.Command("docker", "build", "-t", opts.ImageTag, "-f", opts.Dockerfile, opts.ContextDir)
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}

	// 2) docker create (get a throwaway container)
	create := exec.Command("docker", "create", opts.ImageTag, "sh", "-lc", "echo exporting")
	cidb, err := create.Output()
	if err != nil {
		return fmt.Errorf("docker create: %w", err)
	}
	cid := string(bytes.TrimSpace(cidb))

	// Ensure cleanup of the throwaway container
	defer func() { _ = exec.Command("docker", "rm", cid).Run() }()

	// Prepare output dir
	if err := os.MkdirAll(opts.OutputRootfs, 0o755); err != nil {
		return fmt.Errorf("mkdir rootfs: %w", err)
	}

	// 3) docker export to a temporary tar (avoid piping issues)
	tmpTar, err := os.CreateTemp("", "rootfs-*.tar")
	if err != nil {
		return fmt.Errorf("create temp tar: %w", err)
	}
	tmpTarPath := tmpTar.Name()
	_ = tmpTar.Close()
	defer os.Remove(tmpTarPath)

	export := exec.Command("docker", "export", "-o", tmpTarPath, cid)
	// capture stderr for better diagnostics
	var exportStderr bytes.Buffer
	export.Stderr = &exportStderr
	if err := export.Run(); err != nil {
		msg := exportStderr.String()
		if msg == "" {
			msg = "docker export failed"
		}
		return fmt.Errorf("docker export: %w: %s", err, msg)
	}

	// 4) extract without preserving owner/perms (macOS bind mounts)
	// tar --no-same-owner --no-same-permissions -C <rootfs> -xf <tmpTar>
	tar := exec.Command("tar", "--no-same-owner", "--no-same-permissions", "-C", opts.OutputRootfs, "-xf", tmpTarPath)
	// stream tar stderr to user
	tar.Stdout = io.Discard
	tar.Stderr = os.Stderr
	if err := tar.Run(); err != nil {
		return fmt.Errorf("tar extract: %w", err)
	}

	// 5) done
	// Optional: ensure /proc mountpoint exists inside rootfs; not strictly necessary
	_ = os.MkdirAll(filepath.Join(opts.OutputRootfs, "proc"), 0o555)

	return nil
}
