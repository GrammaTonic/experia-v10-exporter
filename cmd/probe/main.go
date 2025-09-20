package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "http://127.0.0.1:9100/metrics", "URL to probe")
	timeoutStr := flag.String("timeout", "2s", "request timeout")
	flag.Parse()

	dur, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid timeout: %v\n", err)
		os.Exit(2)
	}

	client := &http.Client{Timeout: dur}
	resp, err := client.Get(*url)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		os.Exit(1)
	}
	os.Exit(0)
}
