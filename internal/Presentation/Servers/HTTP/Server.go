package HTTP

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"encoding/json"
	"net/http"
	"strconv"
)

type Server struct {
	service OrdersServices.IGetService
}

func NewServer(service OrdersServices.IGetService) *Server {
	return &Server{service: service}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/" {
		http.NotFound(w, r)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	item, err := s.service.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}
