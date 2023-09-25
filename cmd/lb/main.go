package main

import (
	"log/slog"
	"os"

	"github.com/jboursiquot/loadbalancer"
)

func main() {
	ports := []string{":8081", ":8082", ":8083"}
	lb := loadbalancer.NewLoadBalancer()
	if err := lb.Start(ports); err != nil {
		slog.Error("LoadBalancer failed to start", "error", err.Error())
		os.Exit(1)
	}
}
