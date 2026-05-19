package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sudovishal/vortexlog/internal/api"
	"github.com/sudovishal/vortexlog/internal/worker"
)

func main() {

	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")

	ctx := context.Background()

	// Create a connection pool instead of a single connection
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Verify the connection is working
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to ping database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully connected to database pool!")

	// Initialize the queries object
	// queries := database.New(pool)

	incomingChan := make(chan api.LogPayload, 100)

	// Register HTTP handler
	http.HandleFunc("POST /logs", api.HandleLogs(incomingChan))

	var wg sync.WaitGroup
	worker.StartWorkerPool(incomingChan, &wg, pool)

	// Graceful shutdown
	server := &http.Server{
		Addr: ":3001",
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Println("Server is listening on port 3001...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		}
	}()

	sig := <-quit
	fmt.Printf("Received signal: %v\n", sig)

	// Stop accepting new requests
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server shutdown failed: %v\n", err)
		os.Exit(1)
	}

	// Stop workers
	close(incomingChan)
	fmt.Println("Waiting for workers to finish...")
	wg.Wait()

	// Close DB pool after workers are done
	pool.Close()
	fmt.Println("Server exited gracefully")
}
