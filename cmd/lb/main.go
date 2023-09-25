package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jboursiquot/loadbalancer"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ports := []string{":8081", ":8082", ":8083"}
	lb := loadbalancer.NewLoadBalancer(":8080")
	go func() {
		if err := lb.Start(ports); err != nil {
			slog.Warn("LoadBalancer", "error", err.Error())
			os.Exit(1)
		}
	}()

	slog.Info("Started LoadBalancer", "ports", ports)

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := lb.Stop(ctx); err != nil {
		slog.Warn("LoadBalancer", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("Done")
}
