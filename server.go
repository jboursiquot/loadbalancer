package loadbalancer

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"log/slog"
)

type server struct {
	httpServer http.Server
}

func (s *server) start() error {
	slog.Info("Starting server", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *server) stop(ctx context.Context) error {
	slog.Info("Stopping server", "addr", s.httpServer.Addr)
	return s.httpServer.Shutdown(ctx)
}

func newServer(addr string) *server {
	r := http.NewServeMux()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if rand.Intn(2) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("Server %s handled request %s", addr, r.Header.Get("X-Request-ID"))))
	})

	return &server{
		httpServer: http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}
