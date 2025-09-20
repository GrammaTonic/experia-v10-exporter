package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/collector"
)

func main() {
	ip := net.ParseIP(os.Getenv("EXPERIA_V10_ROUTER_IP"))
	if ip == nil {
		fmt.Fprintln(os.Stderr, "EXPERIA_V10_ROUTER_IP is not set or is invalid")
		os.Exit(2)
	}
	username := os.Getenv("EXPERIA_V10_ROUTER_USERNAME")
	password := os.Getenv("EXPERIA_V10_ROUTER_PASSWORD")
	c := collector.NewCollector(ip, username, password, 10*time.Second)
	if err := c.Login(); err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Login succeeded")
	token := c.SessionToken()
	fmt.Printf("Session token: %s\n", token)

	// Try a few URL variants to inspect cookies stored under different hostnames/paths.
	variants := []string{}
	if b := os.Getenv("EXPERIA_V10_BASE_URL"); b != "" {
		variants = append(variants, b)
	}
	variants = append(variants, "http://"+os.Getenv("EXPERIA_V10_ROUTER_IP"))
	variants = append(variants, "http://"+os.Getenv("EXPERIA_V10_ROUTER_IP")+":80")
	variants = append(variants, "http://"+os.Getenv("EXPERIA_V10_ROUTER_IP")+"/")

	for _, url := range variants {
		cookies := c.CookiesForHost(url)
		fmt.Printf("Found %d cookies for %s\n", len(cookies), url)
		for _, ck := range cookies {
			fmt.Printf("Cookie: %s=%s; Path=%s; Domain=%s; Expires=%v; Secure=%v; HttpOnly=%v\n",
				ck.Name, ck.Value, ck.Path, ck.Domain, ck.Expires, ck.Secure, ck.HttpOnly)
		}
	}
}
