package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/1password/onepassword-sdk-go"
)

func main() {
	// Gets your service account token from the OP_SERVICE_ACCOUNT_TOKEN environment variable.
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	if token == "" {
		log.Fatal("OP_SERVICE_ACCOUNT_TOKEN environment variable is not set")
	}

	// Authenticates with your service account token and connects to 1Password.
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
	)

	// Handles any errors that may have occurred during the authentication process.
	if err != nil {
		log.Fatalf("Error creating 1Password client")
	}

	// Resolves the secret "op://vault/item/field" and prints it to the console.
	secret, err := client.Secrets.Resolve(context.Background(), "op://robot_secrets/Postgres/Sabertooth Admin/password")
	if err != nil {
		log.Fatalf("Error resolving secret: %v", err)
	}

	// Prints the resolved secret to the console.
	fmt.Println("Secret:", secret)
}
