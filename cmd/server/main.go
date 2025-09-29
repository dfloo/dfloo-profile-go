package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dfloo/dfloo-profile-go/internal/router"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := InitPool(ctx); err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer Pool.Close()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
		Pool.Close()
		os.Exit(0)
	}()

	log.Print("Server listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", router.New(Pool))
	if err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

var (
	Pool *pgxpool.Pool
	once sync.Once
)

func InitPool(ctx context.Context) error {
	var err error
	once.Do(func() {
		Pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	})
	return err
}
