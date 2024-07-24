package util

import (
	"fmt"
	"os"
	"os/exec"
)

func Label(label, f string, a ...any) {
	fmt.Printf("%-10s ", label)
	fmt.Printf(f, a...)
	fmt.Println()
}

func RunCommand(name string, a ...string) error {
	cmd := exec.Command(name, a...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
