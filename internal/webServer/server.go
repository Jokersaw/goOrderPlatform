package webserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/jokersaw/goOrderPlatform/internal/cache"
	"github.com/jokersaw/goOrderPlatform/internal/db"
)

type Server struct {
	db          *sqlx.DB
	cache       *cache.OrderCache
	frontendDir string
}

func NewServer(db *sqlx.DB, c *cache.OrderCache, frontendDir string) *Server {
	return &Server{
		db:          db,
		cache:       c,
		frontendDir: frontendDir,
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /order/{id}", s.getOrderHandler)
	mux.Handle("/", http.FileServer(http.Dir(s.frontendDir)))

	log.Printf("HTTP server listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")

	if order, ok := s.cache.Get(orderID); ok {
		writeJSON(w, http.StatusOK, order)
		return
	}

	order, err := db.GetOrder(s.db, orderID)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	s.cache.Set(order)
	log.Printf("cached order: %s", order.OrderUID)

	writeJSON(w, http.StatusOK, order)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: encode error: %v", err)
	}
}
