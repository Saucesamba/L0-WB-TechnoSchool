package handlers

import (
	"L0WB/internal/cache"
	"L0WB/internal/config"
	"L0WB/internal/db"
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type OrderHandler struct {
	*BaseHandler
}

func NewProductHandler(db db.Database, config *config.AppConfig, Cache *cache.Cache) *OrderHandler {
	return &OrderHandler{&BaseHandler{DB: db, Config: config, Cache: Cache}}
}

func (h *OrderHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId := vars["order_id"]
	order, ok := (*h.Cache).Get(orderId)
	if ok {
		ResponseWithJSON(w, http.StatusOK, order)
		log.Println("Get order from cache")
		return
	}
	order, err := h.DB.GetOrder(context.Background(), orderId)
	if err != nil {
		ResponseWithJSON(w, http.StatusNotFound, err.Error())
		return
	}
	log.Println("Get order from DB")
	(*h.Cache).Add(orderId, order)
	ResponseWithJSON(w, http.StatusOK, order)
	return
}
