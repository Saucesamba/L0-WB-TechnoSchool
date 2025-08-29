package handlers

import (
	"L0WB/internal/cache"
	"L0WB/internal/config"
	"L0WB/internal/db"
	"encoding/json"
	"net/http"
)

type BaseHandler struct {
	DB     db.Database
	Config *config.AppConfig
	Cache  *cache.Cache
}

func NewBaseHandler(DB db.Database, Config *config.AppConfig, Cache *cache.Cache) *BaseHandler {
	return &BaseHandler{
		DB:     DB,
		Config: Config,
		Cache:  Cache,
	}
}

func ResponseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
