package main

import (
	"context"
	payments "dev-pay-client/grpcproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var action, username, password string
var newsletter bool

const (
	REQUEST_COUNT = 1000000
	CONCURRENCY   = 100
	BATCH_SIZE    = 100
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

	client := payments.NewCreateAccountServiceClient(conn)
	_ = payments.NewTransactionsLookUpServiceClient(conn)

	var successCount, failCount int64
	startTime := time.Now()

	// Create a channel to distribute work
	jobs := make(chan []uint64, REQUEST_COUNT/BATCH_SIZE)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < CONCURRENCY; i++ {
		wg.Add(1)
		go worker(client, jobs, &successCount, &failCount, &wg)
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

	// Wait for all workers to finish
	wg.Wait()

	totalTime := time.Since(startTime)
	totalSeconds := totalTime.Seconds()
	successPerSecond := float64(successCount) / totalSeconds

	log.Printf("Total requests: %d, Successful: %d, Failed: %d", REQUEST_COUNT, successCount, failCount)
	log.Printf("Total time taken: %.2f seconds", totalSeconds)
	log.Printf("Successful requests per second: %.2f", successPerSecond)

	log.Printf("loading all the records to verify the transaction...\n")
	time.Sleep(3 * time.Second)
	//verifyAccounts(client2, 1, REQUEST_COUNT)

	//// Initial selection for action
	//initialForm := huh.NewForm(
	//	huh.NewGroup(
	//		huh.NewSelect[string]().
	//			Title("What's your plan today?").
	//			Options(
	//				huh.NewOption("Use Account!", "login"),
	//				huh.NewOption("Create Account?", "register"),
	//			).
	//			Value(&action),
	//	),
	//)
	//
	//// Run the initial form
	//err = initialForm.Run()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Depending on the action, we create different forms
	//switch action {
	//case "login":
	//	loginForm := huh.NewForm(
	//		huh.NewGroup(
	//			huh.NewInput().
	//				Title("Enter accountID:").
	//				Value(&username),
	//			huh.NewInput().
	//				Title("Enter password:").
	//				EchoMode(huh.EchoModePassword).
	//				Value(&password),
	//		),
	//	)
	//
	//	err := loginForm.Run()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	uicomponents.NewProgressBar()
	//	uicomponents.NewTransfersTable()
	//
	//	// Show the summary of login information
	//	fmt.Printf("\nSummary:\nAccountID: %s\n", username)
	//
	//case "register":
	//	registerForm := huh.NewForm(
	//		huh.NewGroup(
	//			huh.NewInput().
	//				Title("Choose a accountID:").
	//				Value(&username).
	//				Validate(func(str string) error {
	//					if len(str) < 3 {
	//						return errors.New("accountID must be at least 3 characters long")
	//					}
	//					return nil
	//				}),
	//			huh.NewInput().
	//				Title("Choose a password:").
	//				Password(true).
	//				Value(&password).
	//				Validate(func(str string) error {
	//					if len(str) < 6 {
	//						return errors.New("password must be at least 6 characters long")
	//					}
	//					return nil
	//				}),
	//			huh.NewConfirm().
	//				Title("Would you like to receive our newsletter?").
	//				Value(&newsletter),
	//		),
	//	)
	//
	//	err := registerForm.Run()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	// Show the summary of registration information
	//	fmt.Printf("\nSummary:\nAccountID: %s\nNewsletter Subscription: %t\n", username, newsletter)
	//	fmt.Println("Password: [hidden]")
	//}
	//
	//if action != "login" {
	//	// Optional: Add a confirmation step if needed
	//	confirm := false
	//	confirmationForm := huh.NewForm(
	//		huh.NewGroup(
	//			huh.NewConfirm().
	//				Title("Is the above information correct?").
	//				Value(&confirm),
	//		),
	//	)
	//	err = confirmationForm.Run()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	if !confirm {
	//		fmt.Println("You chose to modify your information. Restarting the form...")
	//		// Optionally, restart the form or allow modifications here
	//	} else {
	//		fmt.Println("Thank you! Your information has been accepted.")
	//	}
	//}

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

func verifyAccounts(client payments.TransactionsLookUpServiceClient, start, end uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req := &payments.LookupAccountsRequest{
		AccountIds: createRange(start, end),
	}

	resp, err := client.LookupAccounts(ctx, req)
	if err != nil {
		log.Printf("Failed to lookup accounts: %v", err)
		return
	}

	log.Printf("Verified %d accounts out of %d requested", len(resp.Accounts), end-start)
}
