package readiness

import (
	"net/http"
	"sync"
)

type Handler struct {
	ready bool
	mutex sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		ready: false,
	}
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.ready {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Not OK"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Set(ready bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.ready = ready
}
