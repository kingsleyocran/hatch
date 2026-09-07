//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
)

func setForkAttrs(child *exec.Cmd) {
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
