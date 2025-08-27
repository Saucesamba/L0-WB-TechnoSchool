package app

import (
	"L0WB/internal/config"
	"L0WB/internal/db"
	"L0WB/internal/kafka"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type App struct {
	Router     *mux.Router
	Config     *config.AppConfig
	DB         db.Database
	HTTPServer *http.Server
	Consumer   *kafka.Consumer
}

func NewApp() *App {
	return &App{}
}

func (app *App) Initialize() error {
	cfg, err := new(config.AppConfig).LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	app.Config = cfg
	//fmt.Println(app.Config)
	log.Println("Waiting for database to start...")
	time.Sleep(10 * time.Second)

	WbDB, err := new(db.WbDB).NewDB(&config.DBConfig{
		Host:     app.Config.Postgres.Host,
		Port:     app.Config.Postgres.Port,
		User:     app.Config.Postgres.User,
		Password: app.Config.Postgres.Password,
		DBName:   app.Config.Postgres.DBName,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	app.DB = WbDB

	kafkaConsumer, err := kafka.NewConsumer(app.Config.Kafka, app.DB)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	app.Consumer = kafkaConsumer

	app.Router = mux.NewRouter()
	app.setRouters()
	app.HTTPServer = &http.Server{
		Addr:    app.Config.HTTP.Host + ":" + app.Config.HTTP.Port,
		Handler: app.Router,
	}
	return nil
}

func (app *App) RunConsumer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Kafka consumer")
			return
		default:
			msg, err := app.Consumer.ReadMessage(ctx)
			if err != nil {
				log.Printf("Failed to read message: %v", err)
				continue
			}
			err = app.Consumer.ProcessMessage(ctx, msg)
			if err != nil {
				log.Printf("Failed to process message: %v", err)
			}
		}
	}
}

func (app *App) setRouters() {
	app.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is running"))
	})
}
