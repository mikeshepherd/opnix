package secrets

import (
    "testing"

	"path/filepath"
	"fmt"
	"os"
    "github.com/brizzbuzz/opnix/internal/config"
)

// Mock client for testing
type mockClient struct {
    secrets map[string]string
}

func (m *mockClient) ResolveSecret(reference string) (string, error) {
    if value, ok := m.secrets[reference]; ok {
        return value, nil
    }
    return "", fmt.Errorf("secret not found")
}

func TestProcessor(t *testing.T) {
    // Create mock client
    mock := &mockClient{
        secrets: map[string]string{
            "op://vault/item/field": "test-secret-value",
        },
    }

    // Create temp output directory
    tmpDir, err := os.MkdirTemp("", "opnix-processor-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)

    // Create processor
    processor := NewProcessor(mock, tmpDir)

    // Create test config
    cfg := &config.Config{
        Secrets: []config.Secret{
            {
                Path:      "test/secret",
                Reference: "op://vault/item/field",
            },
        },
    }

    // Process secrets
    if err := processor.Process(cfg); err != nil {
        t.Fatalf("Failed to process secrets: %v", err)
    }

    // Verify output
    outputPath := filepath.Join(tmpDir, "test/secret")
    content, err := os.ReadFile(outputPath)
    if err != nil {
        t.Fatalf("Failed to read output file: %v", err)
    }

    if string(content) != "test-secret-value" {
        t.Errorf("Expected secret value test-secret-value, got %s", string(content))
    }
}
