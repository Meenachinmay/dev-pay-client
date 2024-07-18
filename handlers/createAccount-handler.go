package handlers

import (
	"context"
	payments "dev-pay-client/grpcproto"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

func (config *Config) CreateAccountHandler(successCount *int64, failCount *int64, wg *sync.WaitGroup) {

	client := payments.NewCreateAccountServiceClient(config.Client)

	// Create a channel to distribute work: Create account
	jobs := make(chan []uint64, REQUEST_COUNT/BATCH_SIZE)

	// Start worker goroutines: Create account
	for i := 0; i < CONCURRENCY; i++ {
		wg.Add(1)
		go worker(client, jobs, successCount, failCount, wg)
	}

	// Send jobs to the channel in batches
	for i := uint64(1); i <= REQUEST_COUNT; i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > REQUEST_COUNT {
			end = REQUEST_COUNT + 1
		}
		jobs <- createRange(i, end)
	}
	close(jobs)
}

func createRange(start, end uint64) []uint64 {
	r := make([]uint64, end-start)
	for i := range r {
		r[i] = start + uint64(i)
	}
	return r
}

func worker(client payments.CreateAccountServiceClient, jobs <-chan []uint64, successCount, failCount *int64, wg *sync.WaitGroup) {
	defer wg.Done()

	for batch := range jobs {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

		accounts := make([]*payments.Account, len(batch))
		for i, id := range batch {
			accounts[i] = &payments.Account{
				Id:     id,
				Ledger: 1,
				Code:   1,
			}
		}

		req := &payments.CreateAccountBatchRequest{Accounts: accounts}
		_, err := client.CreateAccountBatch(ctx, req)

		cancel()

		if err != nil {
			log.Printf("Failed to create batch of accounts: %v", err)
			atomic.AddInt64(failCount, int64(len(batch)))
		} else {
			atomic.AddInt64(successCount, int64(len(batch)))
			log.Printf("Created batch of %d accounts", len(batch))
		}
	}
}
