package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sudovishal/vortexlog/internal/api"
	"github.com/sudovishal/vortexlog/internal/database"
)

// StartWorkerPool starts a pool of goroutines to process logs from the incoming channel
// with batching logic.
func StartWorkerPool(incomingChan <-chan api.LogPayload, wg *sync.WaitGroup, pool *pgxpool.Pool) {
	const numWorkers = 3
	const batchSize = 100
	const batchTimeout = 10 * time.Second

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			batch := make([]api.LogPayload, 0, batchSize)
			ticker := time.NewTicker(batchTimeout)
			defer ticker.Stop()
			queries := database.New(pool)
			for {
				select {
				case log, ok := <-incomingChan:
					if !ok {
						// The channel was closed (useful later for graceful shutdown)
						if len(batch) > 0 {
							fmt.Printf("Worker %d flushing final %d logs before shutdown\n", workerID, len(batch))
							// TODO: Actual Database Insert
							flushBatch(queries, batch, workerID)

						}
						return
					}

					// Add the log to our current batch
					batch = append(batch, log)

					// If the batch is full, flush it immediately
					if len(batch) >= batchSize {
						fmt.Printf("Worker %d batch full! Flushing %d logs to database\n", workerID, len(batch))
						// TODO: Actual Database Insert
						flushBatch(queries, batch, i)

						// Clear the batch for the next round
						batch = make([]api.LogPayload, 0, batchSize)
						// Reset the stopwatch so we don't prematurely trigger the ticker
						ticker.Reset(batchTimeout)
					}

				case <-ticker.C:
					// The 500ms stopwatch went off! Flush whatever we have, even if it's just 1 log.
					if len(batch) > 0 {
						fmt.Printf("Worker %d timer went off! Flushing %d logs to database\n", workerID, len(batch))
						// TODO: Actual Database Insert
						flushBatch(queries, batch, i)

						// Clear the batch for the next round
						batch = make([]api.LogPayload, 0, batchSize)
					}
				}
			}
		}(i)
	}
}

func flushBatch(queries *database.Queries, batch []api.LogPayload, workerID int) {
	if len(batch) == 0 {
		return
	}
	// 1. Create a slice of the struct type that sqlc generated for us
	params := make([]database.CreateLogsParams, len(batch))
	// 2. Loop using the index 'i' so we can assign to the correct slot in our new slice
	for i, log := range batch {
		var traceID pgtype.Text
		if log.TraceID != nil {
			traceID = pgtype.Text{String: *log.TraceID, Valid: true}
		}

		var metadataBytes []byte
		if log.Metadata != nil {
			var err error
			metadataBytes, err = json.Marshal(log.Metadata)
			if err != nil {
				fmt.Printf("Worker %d failed to marshal metadata for log: %v\n", workerID, err)
			}
		}

		// 3. Create the struct using its actual Type Name: database.CreateLogsParams
		params[i] = database.CreateLogsParams{
			ServiceName: log.ServiceName,
			LogLevel:    log.LogLevel,
			Message:     log.Message,
			CreatedAt:   log.CreatedAt,
			TraceID:     traceID,
			Metadata:    metadataBytes,
		}
	}
	// 4. Execute the high-speed Batch Insert!
	ctx := context.Background()
	insertedCount, err := queries.CreateLogs(ctx, params)
	if err != nil {
		fmt.Printf("Worker %d failed to insert batch: %v\n", workerID, err)
		return
	}
	fmt.Printf("Worker %d successfully batch-inserted %d logs!\n", workerID, insertedCount)
}
