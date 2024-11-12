package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoad(t *testing.T) {
    // Create temp config file
    tmpDir, err := os.MkdirTemp("", "opnix-tests-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)

    configPath := filepath.Join(tmpDir, "config.json")
    configData := `{
        "secrets": [
            {
                "path": "test/secret",
                "reference": "op://vault/item/field"
            }
        ]
    }`

    if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
        t.Fatalf("Failed to write config file: %v", err)
    }

    cfg, err := Load(configPath)
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    if len(cfg.Secrets) != 1 {
        t.Errorf("Expected 1 secret, got %d", len(cfg.Secrets))
    }

    if cfg.Secrets[0].Path != "test/secret" {
        t.Errorf("Expected path test/secret, got %s", cfg.Secrets[0].Path)
    }
}
