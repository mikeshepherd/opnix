package main

import (
    "flag"
    "log"

    "github.com/brizzbuzz/opnix/internal/config"
    "github.com/brizzbuzz/opnix/internal/onepass"
    "github.com/brizzbuzz/opnix/internal/secrets"
)

func main() {
    configFile := flag.String("config", "secrets.json", "Path to secrets configuration file")
    outputDir := flag.String("output", "secrets", "Directory to store retrieved secrets")
    tokenFile := flag.String("token-file", "", "Path to file containing 1Password service account token")
    flag.Parse()

    // Load configuration
    cfg, err := config.Load(*configFile)
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Initialize 1Password client
    client, err := onepass.NewClient(*tokenFile)
    if err != nil {
        log.Fatalf("Error creating 1Password client: %v", err)
    }

    // Process secrets
    processor := secrets.NewProcessor(client, *outputDir)
    if err := processor.Process(cfg); err != nil {
        log.Fatalf("Error processing secrets: %v", err)
    }
}
