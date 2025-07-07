package handlers

import (
	"OrdersService/internal/cache"
	"OrdersService/internal/repository"
	"OrdersService/pkg/logging"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type IHandler interface {
	Register(router *httprouter.Router)
}

type Handler struct {
	cache  *cache.OrderCache
	repo   *repository.OrderRepository
	logger *logging.Logger
}

func NewOrderHandler(cache *cache.OrderCache, repo *repository.OrderRepository, logger *logging.Logger) *Handler {
	return &Handler{
		cache:  cache,
		repo:   repo,
		logger: logger,
	}
}

func (h *Handler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	// Try to get from cache first
	if order, ok := h.cache.Get(orderUID); ok {
		h.logger.Infof("Order %s found in cache", orderUID)
		respondWithJSON(w, http.StatusOK, order)
		return
	}

	// If not in cache, try to get from DB
	order, err := h.repo.FindOne(r.Context(), orderUID)
	if err != nil {
		h.logger.Errorf("Failed to get order %s from DB: %v", orderUID, err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if order == nil {
		h.logger.Infof("Order %s not found", orderUID)
		respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Update cache
	h.cache.Set(*order)

	respondWithJSON(w, http.StatusOK, order)
}

func (h *Handler) GetIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/index.html")
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
