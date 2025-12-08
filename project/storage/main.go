package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const gatewayUrl = "http://gateway:9090"

func register(addr string, retries int) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < retries; i++ {
		resp, err := client.Post(gatewayUrl+"/register", "text/plain", bytes.NewBufferString(addr))
		if err == nil && resp.StatusCode == http.StatusCreated {
			log.Printf("Successfully registered with gateway -> storage: %s", addr)
			return true
		}
		log.Printf("Register attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return false
}

func heartbeat(addr string, stop <-chan struct{}) {
	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := client.Post(gatewayUrl+"/heartbeat", "text/plain", bytes.NewBufferString(addr))
			if err != nil {
				log.Printf("Heartbeat failed: %v", err)
			} else {
				resp.Body.Close()
				log.Printf("Heartbeat sent by: %s", addr)
			}
		case <-stop:
			log.Println("Stopping heartbeat")
			return
		}
	}
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"service": os.Getenv("STORAGE_NAME"),
		"time":    time.Now(),
		"data":    []string{os.Getenv("STORAGE_NAME")},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Invoked healthcheck")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}

func main() {
	servicePort := "9999"
	serviceName := os.Getenv("STORAGE_NAME")
	if serviceName == "" {
		serviceName = "storage1"
	}

	storageAddr := "http://" + serviceName + ":" + servicePort

	mux := http.NewServeMux()
	mux.HandleFunc("/data", dataHandler)
	mux.HandleFunc("/healthz", healthHandler)

	server := &http.Server{
		Addr:    ":" + servicePort,
		Handler: mux,
	}

	go func() {
		log.Println("Starting storage service on port " + servicePort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	if !register(storageAddr, 10) {
		log.Println("Failed to register with gateway, exiting")
		os.Exit(1)
	}

	stopHeartbeat := make(chan struct{})
	go heartbeat(storageAddr, stopHeartbeat)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received")

	close(stopHeartbeat)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Storage service stopped gracefully")
}
