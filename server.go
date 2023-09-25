package loadbalancer

import (
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

func newServer(addr string) *server {
	r := http.NewServeMux()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server " + addr))
	})

	return &server{
		httpServer: http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}
