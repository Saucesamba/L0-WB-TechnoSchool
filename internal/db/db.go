package db

import (
	"L0WB/internal/config"
	"L0WB/internal/models"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Database interface {
	NewDB(*config.DBConfig) (Database, error)
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
}

type WbDB struct {
	*sql.DB
}

func (w *WbDB) NewDB(cfg *config.DBConfig) (Database, error) {
	source := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	dbConn, err := sql.Open("postgres", source)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}
	return &WbDB{dbConn}, err
}
