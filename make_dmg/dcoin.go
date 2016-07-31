package main

import (
  "fmt"
  "path/filepath"
  "os/exec"
  "os"
)
func main() {
        dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
        fmt.Println(dir+"/dcoinbin")
        cmd := exec.Command(dir+"/dcoinbin")
        err := cmd.Run()
        if err != nil {
                os.Exit(1)
                fmt.Println("err=",err)
        }
}
