package onepass

import (
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/1password/onepassword-sdk-go"
)

type Client struct {
    op *onepassword.Client
}

func NewClient(tokenFile string) (*Client, error) {
    token, err := getToken(tokenFile)
    if err != nil {
        return nil, err
    }

    client, err := onepassword.NewClient(context.Background(),
        onepassword.WithServiceAccountToken(token),
        onepassword.WithIntegrationInfo("NixOS Secrets Integration", "v1.0.0"),
    )
    if err != nil {
        return nil, err
    }

    return &Client{op: client}, nil
}

func (c *Client) ResolveSecret(reference string) (string, error) {
    return c.op.Secrets.Resolve(context.Background(), reference)
}

func getToken(tokenFile string) (string, error) {
    if token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN"); token != "" {
        return strings.TrimSpace(token), nil
    }

    if tokenFile != "" {
        data, err := os.ReadFile(tokenFile)
        if err != nil {
            return "", fmt.Errorf("failed to read token file: %w", err)
        }
        return strings.TrimSpace(string(data)), nil
    }

    return "", fmt.Errorf("no token provided: set OP_SERVICE_ACCOUNT_TOKEN or provide token file")
}
