package main

import (
	"OrdersService/internal/cache"
	cfg "OrdersService/internal/config"
	"OrdersService/internal/handlers"
	"OrdersService/internal/kafka"
	"OrdersService/internal/repository"
	"OrdersService/pkg/client/postgresql"
	"OrdersService/pkg/logging"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := logging.GetLogger()

	logger.Info("Read configuration")
	config := cfg.GetConfig()

	postgresDB, err := postgresql.NewClient(context.Background(), 5, *config)
	if err != nil {
		logger.Fatalf("Failed to connect database: %v", err)
	}
	defer postgresDB.Close()

	logger.Info("Initialize repository")
	orderRepo := repository.NewOrderRepository(postgresDB, logger)

	logger.Info("Initialize cache")
	orderCache := cache.NewOrderCache()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := orderCache.RestoreFromDB(ctx, orderRepo.FindAll); err != nil {
		logger.Fatalf("Failed to restore cache from database: %v", err)
	}

	kafkaConsumer := kafka.NewConsumer(
		config,
		orderRepo,
		orderCache,
		logger,
	)

	kafkaConsumer.Start()
	defer kafkaConsumer.Stop()

	startHTTPServer(&config.HTTP, orderCache, orderRepo, logger)
}

func startHTTPServer(httpCfg *cfg.HTTPConfig, cache *cache.OrderCache, repo *repository.OrderRepository, logger *logging.Logger) {
	orderHandler := handlers.NewOrderHandler(cache, repo, logger)

	router := mux.NewRouter()
	router.HandleFunc("/order/{order_uid}", orderHandler.GetOrderByUID).Methods("GET")
	router.HandleFunc("/", orderHandler.GetIndex).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	server := &http.Server{
		Addr:         ":" + httpCfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Infof("Starting HTTP server on port %s", httpCfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	waitForShutdown(server, logger)
}

func waitForShutdown(server *http.Server, logger *logging.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited properly")
}
