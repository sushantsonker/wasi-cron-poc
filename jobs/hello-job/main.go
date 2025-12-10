package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	jobName := os.Getenv("JOB_NAME")
	if jobName == "" {
		jobName = "hello-job"
	}

	now := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] Hello from WASM job at %s\n", jobName, now)
}
