package main

import (
	"dev-pay-client/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	configData, err := os.ReadFile("grpc_config.json")
	if err != nil {
		log.Fatalf("failed to read gRPC config file: %v", err)
	}

	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(string(configData)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)), // 100MB
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100*1024*1024)), // 100MB
	)
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to gRPC server")

	gRPC := &handlers.Config{
		Client: conn,
	}

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	var successCount, failCount int64
	startTime := time.Now()

	// CALL gRPC SERVICE HERE
	//gRPC.CreateTransferHandler(&successCount, &failCount, &wg) // create transfer service
	gRPC.CreateAccountHandler(&successCount, &failCount, &wg) // create account service

	// Wait for all workers to finish
	wg.Wait()

	totalTime := time.Since(startTime)
	totalSeconds := totalTime.Seconds()
	successPerSecond := float64(successCount) / totalSeconds

	log.Printf("Total requests: %d, Successful: %d, Failed: %d", handlers.REQUEST_COUNT, successCount, failCount)
	log.Printf("Total time taken: %.2f seconds", totalSeconds)
	log.Printf("Successful requests per second: %.2f", successPerSecond)
	log.Printf("loading all the records to verify the transaction...\n")

}
