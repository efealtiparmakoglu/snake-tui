package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Raw terminal mode via `stty`: portable across Linux and macOS without
// cgo or platform-specific ioctls. Returns a restore func.
func enableRawMode() func() {
	var ttyFlag string
	if runtime.GOOS == "darwin" {
		ttyFlag = "-f"
	} else {
		ttyFlag = "-F"
	}
	saved, err := exec.Command("stty", ttyFlag, "/dev/tty", "-g").Output()
	if err != nil {
		return func() {} // no tty: run without raw input
	}
	_ = exec.Command("stty", ttyFlag, "/dev/tty", "raw", "-echo").Run()
	return func() {
		_ = exec.Command("stty", ttyFlag, "/dev/tty", string(saved)).Run()
		fmt.Fprint(os.Stdout, "\033[?25h") // cursor back on
	}
}
