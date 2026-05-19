package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// 1. Define the Log Format Mirror
type LogPayload struct {
	ServiceName string    `json:"service_name"`
	LogLevel    string    `json:"log_level"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

var logLevels = []string{"INFO", "INFO", "INFO", "WARN", "ERROR"}
var messages = []string{
	"User logged in successfully",
	"Failed to connect to database",
	"Payment processed",
	"Cache miss",
	"Disk space running low",
}

// 2. Create the "Single Shot" Worker
func simulateService(serviceName string) {
	client := &http.Client{
		Timeout: 2 * time.Second, // Don't hang forever if the API is stuck
	}
	url := "http://localhost:3001/logs"

	for {
		// Generate random log
		level := logLevels[rand.Intn(len(logLevels))]
		msg := messages[rand.Intn(len(messages))]

		payload := LogPayload{
			ServiceName: serviceName,
			LogLevel:    level,
			Message:     msg,
			CreatedAt:   time.Now(),
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("[%s] Error marshaling JSON: %v\n", serviceName, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Send POST request
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))

		// 3. Add the "Resiliency" Catch (The Retry Loop)
		if err != nil {
			fmt.Printf("[Generator - %s] Target down. Retrying in 1 second... (%v)\n", serviceName, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Check the status code to see if the server rejected it
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Printf("[Generator - %s] Unexpected status code: %d. Retrying...\n", serviceName, resp.StatusCode)
			resp.Body.Close()
			time.Sleep(1 * time.Second)
			continue
		}

		// Close the body immediately to free up the connection
		resp.Body.Close()

		// A tiny sleep to prevent completely destroying our own CPU
		// Adjust this if you want it to go faster or slower!
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	fmt.Println(" Starting Log Generator...")

	// 4. The Concurrency Trigger
	services := []string{"auth-api", "payment-processor", "frontend-bff", "inventory-db"}

	for _, name := range services {
		fmt.Printf("Spinning up mock service: %s\n", name)
		go simulateService(name)
	}

	fmt.Println("All mock services are firing logs at the API!")
	fmt.Println("Press Ctrl+C to stop.")

	// Prevent the main program from exiting instantly
	select {}
}
