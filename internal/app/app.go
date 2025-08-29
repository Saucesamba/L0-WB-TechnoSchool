package app

import (
	"L0WB/internal/cache"
	"L0WB/internal/config"
	"L0WB/internal/db"
	"L0WB/internal/handlers"
	"L0WB/internal/kafka"
	"L0WB/internal/models"
	"context"
	"encoding/json"
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
	Cache      cache.Cache
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
	time.Sleep(3 * time.Second)

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
	} else {
		log.Println("Consumer created")
	}
	app.Consumer = kafkaConsumer

	appCache := cache.NewLRUCache(100)
	orders, err := app.loadOrdersFromDB(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load orders from db: %w", err)
	}
	appCache.LoadAll(orders)
	app.Cache = appCache

	app.Router = mux.NewRouter()
	app.setRouters()
	app.HTTPServer = &http.Server{
		Addr:    app.Config.HTTP.Host + ":" + app.Config.HTTP.Port,
		Handler: app.Router,
	}
	return nil
}

func (app *App) Start() error {
	go app.RunConsumer(context.Background())

	log.Printf("Starting HTTP server on %s\n", app.HTTPServer.Addr)
	if err := app.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
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
			} else {
				var order models.Order
				err = json.Unmarshal(msg.Value, &order)
				if err == nil {
					app.Cache.Add(order.OrderUID, &order)
				}
			}
		}
	}
}

func (app *App) setRouters() {
	app.Router.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./front/index.html")
	})
	fs := http.FileServer(http.Dir("./front/"))
	app.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	handler := handlers.NewProductHandler(app.DB, app.Config, &app.Cache)
	app.Router.HandleFunc("/order/{order_id}", handler.GetProduct).Methods("GET")
}

func (app *App) loadOrdersFromDB(ctx context.Context) ([]*models.Order, error) {
	orders, err := app.DB.GetLastOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders from db: %w", err)
	}
	return orders, nil
}
