package main

import (
    "fmt"
    "os/exec"
    "time"
)

func main() {
    start := time.Now()
    cmd := exec.Command("./load_test.sh")
    out, err := cmd.CombinedOutput()
    fmt.Printf("Duration: %v err: %v\n%s\n", time.Since(start), err, out)
}
