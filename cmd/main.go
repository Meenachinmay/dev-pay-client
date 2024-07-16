package main

import (
	"context"
	payments "dev-pay-client/grpcproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync/atomic"
	"time"
)

var action, username, password string
var newsletter bool
var REQUEST_COUNT uint64 = 1000000

func main() {
	configData, err := os.ReadFile("grpc_config.json")
	if err != nil {
		log.Fatal("failed to read gRPC config file: %v", err)
	}

	//
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(string(configData)),
	)
	if err != nil {
		log.Fatalf("failed to create gRPC client, failed to connect gRPC server: %v", err)
	} else {
		log.Println("Connected to gRPC server")
	}
	defer conn.Close()

	client := payments.NewCreateAccountServiceClient(conn)

	// simulate
	var successCount, failCount int64
	startTime := time.Now()
	for i := uint64(1); i <= REQUEST_COUNT; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		account := &payments.Account{
			Id:     i,
			Ledger: 1,
			Code:   1,
		}

		req := &payments.CreateAccountRequest{Account: account}
		resp, err := client.CreateAccount(ctx, req)
		if err != nil {
			log.Printf("Failed to create account ID %d: %v", i, err)
			atomic.AddInt64(&failCount, 1)
			continue
		} else {
			atomic.AddInt64(&successCount, 1)
		}

		log.Printf("Created account ID %d, response: %s", i, resp.Results)
	}

	totalTime := time.Since(startTime)
	totalSeconds := totalTime.Seconds()
	successPerSecond := float64(successCount) / totalSeconds

	log.Printf("Total requests: %d, Successful: %d, Failed: %d", REQUEST_COUNT, successCount, failCount)
	log.Printf("Total time taken: %.2f seconds", totalSeconds)
	log.Printf("Successful requests per second: %.2f", successPerSecond)

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
