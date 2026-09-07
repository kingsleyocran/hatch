//go:build windows

package cmd

import "os/exec"

func setForkAttrs(child *exec.Cmd) {
	// Windows doesn't support Setsid; process is already detached via Start()
}
