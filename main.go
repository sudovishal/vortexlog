package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/sudovishal/vortexlog/internal/api"
)

func main() {

	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Initialize the queries object
	// queries := database.New(conn)

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
