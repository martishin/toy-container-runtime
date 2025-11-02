package container

import (
	"os"
	"path/filepath"
	"strconv"
)

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func writeToFile(path, data string) {
	_ = os.WriteFile(path, []byte(data), 0o644)
}

func SetupPidsCgroup(pid, max int, name string) {
	if setupCgroupV1(pid, max, name) || setupCgroupV2(pid, max, name) {
		return
	}
}

func setupCgroupV1(pid, max int, name string) bool {
	root := "/sys/fs/cgroup/pids"

	if !fileExists(root) {
		return false
	}

	dir := filepath.Join(root, name)
	_ = os.MkdirAll(dir, 0o755)

	writeToFile(filepath.Join(dir, "pids.max"), strconv.Itoa(max))
	writeToFile(filepath.Join(dir, "notify_on_release"), "1")
	writeToFile(filepath.Join(dir, "cgroup.procs"), strconv.Itoa(pid))

	return true
}

func setupCgroupV2(pid, max int, name string) bool {
	root := "/sys/fs/cgroup"

	if !fileExists(filepath.Join(root, "cgroup.controllers")) {
		return false
	}

	dir := filepath.Join(root, name)
	_ = os.MkdirAll(dir, 0o755)

	// Enable +pids on the parent if possible (ignore errors on locked roots)
	if fileExists(filepath.Join(root, "cgroup.subtree_control")) {
		writeToFile(filepath.Join(root, "cgroup.subtree_control"), "+pids")
	}
	if fileExists(filepath.Join(dir, "pids.max")) {
		writeToFile(filepath.Join(dir, "pids.max"), strconv.Itoa(max))
	}
	// Add our process to the **created** cgroup dir, not the root
	if fileExists(filepath.Join(dir, "cgroup.procs")) {
		writeToFile(filepath.Join(dir, "cgroup.procs"), strconv.Itoa(pid))
	}

	return true
}
