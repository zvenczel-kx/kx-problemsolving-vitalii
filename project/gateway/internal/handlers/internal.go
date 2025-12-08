package handlers

import (
	"io"
	"log"
	"net/http"
)

type InternalHandlers struct {
	Registry interface {
		Register(string)
		Heartbeat(string)
	}
}

func (h *InternalHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	addr := string(body)
	if addr == "" {
		http.Error(w, "missing address", http.StatusBadRequest)
		return
	}
	h.Registry.Register(addr)
	log.Printf("Registered: %s", addr)
	w.WriteHeader(http.StatusCreated)
}

func (h *InternalHandlers) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	addr := string(body)
	if addr == "" {
		http.Error(w, "missing address", http.StatusBadRequest)
		return
	}
	h.Registry.Heartbeat(addr)
	log.Printf("Heartbeat received: %s", addr)
	w.WriteHeader(http.StatusOK)
}

func (h *InternalHandlers) Health(w http.ResponseWriter, r *http.Request) {
	log.Println("Invoked healthcheck")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}
