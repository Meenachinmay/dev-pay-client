package main

import (
	"dev-pay-client/uicomponents"
	"errors"
	"fmt"
	"log"

	"github.com/charmbracelet/huh" // Assuming 'huh' is a fictional form library
)

func main() {
	var action, username, password string
	var newsletter bool

	// Initial selection for action
	initialForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What's your plan today?").
				Options(
					huh.NewOption("Use Account!", "login"),
					huh.NewOption("Create Account?", "register"),
				).
				Value(&action),
		),
	)

	// Run the initial form
	err := initialForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Depending on the action, we create different forms
	switch action {
	case "login":
		loginForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter accountID:").
					Value(&username),
				huh.NewInput().
					Title("Enter password:").
					EchoMode(huh.EchoModePassword).
					Value(&password),
			),
		)

		err := loginForm.Run()
		if err != nil {
			log.Fatal(err)
		}

		uicomponents.NewProgressBar()
		// Show the summary of login information
		fmt.Printf("\nSummary:\nAccountID: %s\n", username)

	case "register":
		registerForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Choose a accountID:").
					Value(&username).
					Validate(func(str string) error {
						if len(str) < 3 {
							return errors.New("accountID must be at least 3 characters long")
						}
						return nil
					}),
				huh.NewInput().
					Title("Choose a password:").
					Password(true).
					Value(&password).
					Validate(func(str string) error {
						if len(str) < 6 {
							return errors.New("password must be at least 6 characters long")
						}
						return nil
					}),
				huh.NewConfirm().
					Title("Would you like to receive our newsletter?").
					Value(&newsletter),
			),
		)

		err := registerForm.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Show the summary of registration information
		fmt.Printf("\nSummary:\nAccountID: %s\nNewsletter Subscription: %t\n", username, newsletter)
		fmt.Println("Password: [hidden]")
	}

	if action != "login" {
		// Optional: Add a confirmation step if needed
		confirm := false
		confirmationForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Is the above information correct?").
					Value(&confirm),
			),
		)
		err = confirmationForm.Run()
		if err != nil {
			log.Fatal(err)
		}

		if !confirm {
			fmt.Println("You chose to modify your information. Restarting the form...")
			// Optionally, restart the form or allow modifications here
		} else {
			fmt.Println("Thank you! Your information has been accepted.")
		}
	}

}
