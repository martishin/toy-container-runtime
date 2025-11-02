package container

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	hostname    = "container"
	defaultPids = 20
	cgName      = "demo"
)

func Run(rootfs string, argv []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"child", rootfs}, argv...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = CloneAttrs()
	return cmd.Run()
}

func RunChild(rootfs string, argv []string) error {
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot(%s): %w", rootfs, err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir(/): %w", err)
	}

	if err := makeMountsPrivate(); err != nil {
		return fmt.Errorf("makeMountsPrivate: %w", err)
	}

	if err := mountProcAndTmpfs(); err != nil {
		return fmt.Errorf("mountProcAndTmpfs: %w", err)
	}
	defer unmountProcAndTmpfs()

	SetupPidsCgroup(os.Getpid(), defaultPids, cgName)

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode())
		}
		return fmt.Errorf("exec %q: %w", argv[0], err)
	}
	return nil
}
