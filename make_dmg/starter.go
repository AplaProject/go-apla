package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir + "/aplabin")
	cmd := exec.Command(dir + "/aplabin")
	err := cmd.Run()
	if err != nil {
		fmt.Println("err=", err)
		os.Exit(1)
	}
}
