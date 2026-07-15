package memory

import (
	"net/http"

	"github.com/emrecanterzi/wisp/internal/api"
)

type Handler struct {
	server *api.API
	memory *Memory
}

func NewHandler(server *api.API, memory *Memory) *Handler {
	return &Handler{
		memory: memory,
		server: server,
	}
}

func (h *Handler) RegisterHandlers() {
	h.server.RegisterFuncHandler("GET /", h.GetHandler)
	h.server.RegisterFuncHandler("POST /", h.InsertHandler)
	h.server.RegisterFuncHandler("DELETE /", h.DeleteHandler)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	value, found := h.memory.Get(key)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Key not found"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (h *Handler) InsertHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	err := h.memory.Insert(key, value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return

	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Key-Value pair saved successfully"))
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	deleted, err := h.memory.Delete(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return

	}
	if !deleted {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Key not found"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Key deleted successfully"))
}
