//go:build linux

package container

import "syscall"

func CloneAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
		Pdeathsig:    syscall.SIGKILL,
	}
}
