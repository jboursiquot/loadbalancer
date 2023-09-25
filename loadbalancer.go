package loadbalancer

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type LoadBalancer struct {
	servers         []*server
	roundRobinIndex int
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := &url.URL{Scheme: "http", Host: lb.nextAvailableServer().httpServer.Addr}
	httputil.NewSingleHostReverseProxy(u).ServeHTTP(w, r)
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{}
}

func (lb *LoadBalancer) Start(ports []string) error {
	for _, port := range ports {
		s := newServer(port)
		lb.servers = append(lb.servers, s)
		go func() {
			port := s.httpServer.Addr
			if err := s.start(); err != nil {
				slog.Error("LoadBalancer failed to start server", "port", port, "error", err.Error())
			}
		}()
	}

	r := http.NewServeMux()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lb.ServeHTTP(w, r)
	})

	s := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	slog.Info("LoadBalancer running...", "port", s.Addr)
	return s.ListenAndServe()
}

func (lb *LoadBalancer) Stop() error {
	return nil
}

func (lb *LoadBalancer) nextAvailableServer() *server {
	s := lb.servers[lb.roundRobinIndex]
	lb.roundRobinIndex = (lb.roundRobinIndex + 1) % len(lb.servers)
	return s
}
