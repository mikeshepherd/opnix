package onepass

import (
	"strings"
    "context"
    "fmt"
    "os"

    "github.com/1password/onepassword-sdk-go"
)

type Client struct {
    client *onepassword.Client
}

func NewClient(tokenFile string) (*Client, error) {
    var token string

    if tokenFile != "" {
        data, err := os.ReadFile(tokenFile)
        if err != nil {
            return nil, fmt.Errorf("failed to read token file: %w", err)
        }
        token = strings.TrimSpace(string(data))
    }

    if token == "" {
        return nil, fmt.Errorf("no token provided")
    }

    client, err := onepassword.NewClient(
        context.Background(),
        onepassword.WithServiceAccountToken(token),
        onepassword.WithIntegrationInfo("NixOS Secrets Integration", "v1.0.0"),
    )
    if err != nil {
        return nil, fmt.Errorf("error initializing client: %w", err)
    }

    return &Client{client: client}, nil
}

func (c *Client) ResolveSecret(reference string) (string, error) {
    return c.client.Secrets.Resolve(context.Background(), reference)
}
