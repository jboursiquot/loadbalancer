package loadbalancer

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type LoadBalancer struct {
	servers         []*server
	server          *http.Server
	roundRobinIndex int
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.New().String()
	}

	u := &url.URL{Scheme: "http", Host: lb.nextAvailableServer().httpServer.Addr}
	slog.Info("LoadBalancer", "request", reqID, "server", u.String())

	p := httputil.NewSingleHostReverseProxy(u)
	d := p.Director
	p.Director = func(r *http.Request) {
		d(r)
		r.Header.Set("X-Request-ID", reqID)
	}
	p.ServeHTTP(w, r)
}

func NewLoadBalancer(port string) *LoadBalancer {
	return &LoadBalancer{
		server: &http.Server{
			Addr: port,
		},
	}
}

func (lb *LoadBalancer) Start(ports []string) error {
	for _, port := range ports {
		s := newServer(port)
		lb.servers = append(lb.servers, s)
		go func() {
			port := s.httpServer.Addr
			if err := s.start(); err != nil {
				slog.Warn("Server", "port", port, "error", err.Error())
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
	lb.server.Handler = r

	slog.Info("LoadBalancer running...", "port", lb.server.Addr)
	return lb.server.ListenAndServe()
}

func (lb *LoadBalancer) Stop(ctx context.Context) error {
	// Stop all servers
	g := errgroup.Group{}
	for _, s := range lb.servers {
		s := s
		g.Go(func() error {
			return s.stop(ctx)
		})
	}

	// Wait for all servers to stop
	if err := g.Wait(); err != nil {
		return err
	}

	// Shutdown the load balancer's http server
	if err := lb.server.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("LoadBalancer stopped gracefully", "port", lb.server.Addr)
	return nil
}

func (lb *LoadBalancer) nextAvailableServer() *server {
	s := lb.servers[lb.roundRobinIndex]
	lb.roundRobinIndex = (lb.roundRobinIndex + 1) % len(lb.servers)
	r, err := http.Head("http://" + s.httpServer.Addr + "/health")
	if err != nil || r.StatusCode != http.StatusOK {
		slog.Warn("LoadBalancer skipping unhealthy server", "port", s.httpServer.Addr)
		return lb.nextAvailableServer()
	}
	return s
}
