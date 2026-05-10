package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sudovishal/vortexlog/internal/api"
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

	go func() {
		// This loop runs forever in the background, pulling items off the conveyor belt
		for log := range incomingChan {
			fmt.Printf("Worker pulled a log off the channel: %+v\n", log.Message)
		}
	}()

	fmt.Println("Server is listening on port 3001...")
	if err := http.ListenAndServe(":3001", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
	}
}
