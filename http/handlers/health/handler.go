package health

import (
	"net/http"
	"sync"
)

type Handler struct {
	health bool
	mutex  sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		health: true,
	}
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.health {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Not OK"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Set(health bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.health = health
}
