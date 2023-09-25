package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type server struct {
	proxy *httputil.ReverseProxy
}

func (s *server) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func newServer(targetURL *url.URL) *server {
	return &server{
		proxy: httputil.NewSingleHostReverseProxy(targetURL),
	}
}

type loadBalancer struct {
	port            string
	servers         []*server
	roundRobinIndex int
}

func newLoadBalancer(port string, servers []*server) *loadBalancer {
	return &loadBalancer{
		port:    port,
		servers: servers,
	}
}

func (lb *loadBalancer) nextAvailableServer() *server {
	next := lb.servers[lb.roundRobinIndex]
	lb.roundRobinIndex = (lb.roundRobinIndex + 1) % len(lb.servers)
	return next
}

func main() {
	servers := []*server{
		newServer(&url.URL{Scheme: "https", Host: "go.dev"}),
		newServer(&url.URL{Scheme: "https", Host: "pkg.go.dev"}),
		newServer(&url.URL{Scheme: "https", Host: "vuln.go.dev"}),
	}

	lb := newLoadBalancer(":8080", servers)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.nextAvailableServer()
		server.Serve(w, r)
	})

	slog.Info("Load Balancer started", "port", lb.port)
	if err := http.ListenAndServe(lb.port, nil); err != nil {
		slog.Error("Load Balancer failed to start: %v", err)
		os.Exit(1)
	}
}
