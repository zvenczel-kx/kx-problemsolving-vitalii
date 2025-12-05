package main

import (
	"gateway/internal/handlers"
	"gateway/internal/httpserver"
	"gateway/internal/registry"
	"gateway/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	registry := registry.NewServiceRegistry(15 * time.Second)

	gateway := &services.Gateway{
		Client:      &http.Client{Timeout: 3 * time.Second},
		Registry:    registry,
		MaxRetries:  3,
		Timeout:     2 * time.Second,
		BackoffBase: 200 * time.Millisecond,
	}

	publicHandlers := &handlers.PublicHandlers{Gateway: gateway, Registry: registry}
	internalHandlers := &handlers.InternalHandlers{Registry: registry}

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/data", publicHandlers.Data)
	publicMux.HandleFunc("/status", publicHandlers.Status)

	internalMux := http.NewServeMux()
	internalMux.HandleFunc("/register", internalHandlers.Register)
	internalMux.HandleFunc("/heartbeat", internalHandlers.Heartbeat)
	internalMux.HandleFunc("/healthz", internalHandlers.Health)

	publicServer := &http.Server{Addr: ":8080", Handler: publicMux}
	internalServer := &http.Server{Addr: ":9090", Handler: internalMux}

	stop := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		close(stop)
	}()

	log.Println("starting servers")
	httpserver.StartServers(publicServer, internalServer, stop)
}
