package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// healthURL derives the /api/health URL from a listen address like ":3000".
func healthURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host, port = addr, "3000"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/api/health"
}

// runHealthcheck probes url and returns nil only on a 200 response. It backs the
// container HEALTHCHECK: the same static binary can probe itself, which the
// shell-less distroless runtime image otherwise can't do.
func runHealthcheck(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
