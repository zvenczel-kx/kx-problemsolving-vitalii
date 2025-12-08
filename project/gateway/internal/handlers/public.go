package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"gateway/internal/registry"
	"gateway/internal/services"
	"io"
	"log"
	"net"
	"net/http"
)

type PublicHandlers struct {
	Gateway  *services.Gateway
	Registry interface {
		Status() []registry.ServiceStatus
	}
}

func (h *PublicHandlers) Data(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Gateway.Forward(r.Context())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			http.Error(w, "Timeout", http.StatusServiceUnavailable)
			return
		}
		log.Print("Service Unavailable: " + err.Error())
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)

	_, copyErr := io.Copy(w, resp.Body)
	if copyErr != nil {
		if errors.Is(copyErr, context.Canceled) {
			return
		}
		var netErr *net.OpError
		if errors.As(copyErr, &netErr) {
			return
		}
		log.Printf("copy error: %v", copyErr)
	}
}

func (h *PublicHandlers) Status(w http.ResponseWriter, r *http.Request) {
	list := h.Registry.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
