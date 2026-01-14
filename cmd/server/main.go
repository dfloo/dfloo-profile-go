package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dfloo/dfloo-profile-go/internal/router"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := InitPool(ctx); err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer Pool.Close()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router.New(Pool),
	}

	go func() {
		log.Print("Server listening on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutdown signal received, initiating graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}

var (
	Pool *pgxpool.Pool
	once sync.Once
)

func InitPool(ctx context.Context) error {
	var err error
	once.Do(func() {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("PGHOST"),
			os.Getenv("PGPORT"),
			os.Getenv("PGUSER"),
			os.Getenv("PGPASSWORD"),
			os.Getenv("PGDATABASE"),
		)
		Pool, err = pgxpool.New(ctx, dsn)
	})
	return err
}
