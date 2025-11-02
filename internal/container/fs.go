package container

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func makeMountsPrivate() error {
	if err := syscall.Mount("/", "/", "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		if !errors.Is(err, syscall.EBUSY) {
			return fmt.Errorf("bind-mount /: %w", err)
		}
	}

	_ = syscall.Mount("", "/", "", syscall.MS_SLAVE|syscall.MS_REC, "")

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		if !errors.Is(err, syscall.EINVAL) {
			return fmt.Errorf("set MS_PRIVATE: %w", err)
		}
	}

	return nil
}

func mountProcAndTmpfs() error {
	_ = os.MkdirAll("/proc", 0o555)
	_ = os.MkdirAll("/mytemp", 0o755)

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount /proc: %w", err)
	}
	if err := syscall.Mount("tmpfs", "/mytemp", "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("mount /mytemp: %w", err)
	}
	return nil
}

func unmountProcAndTmpfs() {
	_ = syscall.Unmount("/proc", 0)
	_ = syscall.Unmount("/mytemp", 0)
}
