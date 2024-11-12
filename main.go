package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/1password/onepassword-sdk-go"
)

type Secret struct {
    Path      string `json:"path"`      // Output file path
    Reference string `json:"reference"` // 1Password reference (op://vault/item/field)
}

type Config struct {
    Secrets []Secret `json:"secrets"`
}

func getToken(tokenFile string) (string, error) {
    // First check if token is provided via environment variable
    if token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN"); token != "" {
        return strings.TrimSpace(token), nil
    }

    // If not, try to read from token file if provided
    if tokenFile != "" {
        data, err := os.ReadFile(tokenFile)
        if err != nil {
            return "", fmt.Errorf("failed to read token file: %w", err)
        }
        // Ensure we trim any whitespace or newlines
        return strings.TrimSpace(string(data)), nil
    }

    return "", fmt.Errorf("no token provided: set OP_SERVICE_ACCOUNT_TOKEN or provide token file")
}

func main() {
    configFile := flag.String("config", "secrets.json", "Path to secrets configuration file")
    outputDir := flag.String("output", "secrets", "Directory to store retrieved secrets")
    tokenFile := flag.String("token-file", "", "Path to file containing 1Password service account token")
    flag.Parse()

    // Get token either from env or file
    token, err := getToken(*tokenFile)
    if err != nil {
        log.Fatal(err)
    }

    // Log token format for debugging (first few characters)
    if len(token) > 10 {
        log.Printf("Token prefix: %s...", token[:10])
    }

    // Initialize 1Password client
    client, err := onepassword.NewClient(context.Background(),
        onepassword.WithServiceAccountToken(token),
        onepassword.WithIntegrationInfo("NixOS Secrets Integration", "v1.0.0"),
    )
    if err != nil {
        log.Fatalf("Error creating 1Password client: %v", err)
    }

    // Read config file
    configData, err := os.ReadFile(*configFile)
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    var config Config
    if err := json.Unmarshal(configData, &config); err != nil {
        log.Fatalf("Error parsing config file: %v", err)
    }

    // Create output directory if it doesn't exist
    if err := os.MkdirAll(*outputDir, 0755); err != nil {
        log.Fatalf("Error creating output directory: %v", err)
    }

    // Process each secret
    for _, secret := range config.Secrets {
        // Resolve the secret
        value, err := client.Secrets.Resolve(context.Background(), secret.Reference)
        if err != nil {
            log.Printf("Error resolving secret %s: %v", secret.Reference, err)
            continue
        }

        // Create the full output path
        outputPath := filepath.Join(*outputDir, secret.Path)

        // Create parent directory if it doesn't exist
        if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
            log.Printf("Error creating directory for %s: %v", outputPath, err)
            continue
        }

        // Write the secret to file
        if err := os.WriteFile(outputPath, []byte(value), 0600); err != nil {
            log.Printf("Error writing secret to %s: %v", outputPath, err)
        }
    }
}
