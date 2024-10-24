package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/sahandPgr/Load-balancer/config"
)

// Define the LoadBalancer struct
type LoadBalancer struct {
	Current int
	Mutex   sync.Mutex
}

// Define the Server struct
type Server struct {
	URL       *url.URL
	IsHealthy bool
	Mutex     sync.Mutex
}

// Define the NewLoadBalancer function
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		Current: 0,
	}
}

// Define the getNextServer function
func (loadBalancer *LoadBalancer) getNextServer(servers []*Server) *Server {
	loadBalancer.Mutex.Lock()
	defer loadBalancer.Mutex.Unlock()

	for i := 0; i < len(servers); i++ {
		indexNext := loadBalancer.Current % len(servers)
		nextServer := servers[indexNext]
		loadBalancer.Current++

		nextServer.Mutex.Lock()
		isHealthy := nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}

	}
	return nil
}

// Define the healthCheck function
func healthCheck(server *Server, healthCheckInterval time.Duration) {
	for range time.Tick(healthCheckInterval) {
		res, err := http.Head(server.URL.String())
		server.Mutex.Lock()
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Printf("%s is down\n", server.URL)
			server.IsHealthy = false
		} else {
			server.IsHealthy = true
		}
		server.Mutex.Unlock()
	}
}

// Define the ReverseProxy function
func (server *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(server.URL)
}

func main() {
	config := config.LoadConfig()
	healthCheckInterval, err := time.ParseDuration(config.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Error parsing health check interval: %v", err)
	}

	var servers []*Server
	for _, serverUrl := range config.Servers {
		url, _ := url.Parse(serverUrl)
		server := &Server{URL: url, IsHealthy: true}
		servers = append(servers, server)
		go healthCheck(server, healthCheckInterval)
	}

	loadBalancer := NewLoadBalancer()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := loadBalancer.getNextServer(servers)
		if server == nil {
			http.Error(w, "No healthy servers available", http.StatusServiceUnavailable)
			return
		}
		server.ReverseProxy().ServeHTTP(w, r)
	})

	log.Println("Server listening on port: ", config.Port)
	err = http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
