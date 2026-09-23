package onepass

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	connect "github.com/1Password/connect-sdk-go/connect"
	connectonepassword "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/1password/onepassword-sdk-go"
	"github.com/brizzbuzz/opnix/internal/errors"
)

const (
	defaultInitAttempts    = 3
	defaultResolveAttempts = 3
)

type sdkSecretsAPI interface {
	Resolve(context.Context, string) (string, error)
	ResolveAll(context.Context, []string) (onepassword.ResolveAllResponse, error)
}

type connectSecretsAPI interface {
	GetItem(itemQuery, vaultQuery string) (*connectonepassword.Item, error)
}

type Client struct {
	secrets        sdkSecretsAPI
	connectSecrets connectSecretsAPI
}

var (
	newSDKSecrets = func(ctx context.Context, token string) (sdkSecretsAPI, error) {
		client, err := onepassword.NewClient(
			ctx,
			onepassword.WithServiceAccountToken(token),
			onepassword.WithIntegrationInfo("NixOS Secrets Integration", "v1.0.0"),
		)
		if err != nil {
			return nil, err
		}

		return client.Secrets(), nil
	}
	newConnectSecrets = func(host, token string) connectSecretsAPI {
		return connect.NewClientWithUserAgent(host, token, "opnix/0.10.1")
	}
	retrySleep = time.Sleep
)

// GetToken retrieves token from environment or file
func GetToken(tokenFile string) (string, error) {
	return getToken(tokenFile, "OP_SERVICE_ACCOUNT_TOKEN")
}

func getToken(tokenFile, environmentVariable string) (string, error) {
	// First try environment variable
	if token := os.Getenv(environmentVariable); token != "" {
		return token, nil
	}

	// Then try token file
	if tokenFile != "" {
		data, err := os.ReadFile(tokenFile)
		if err != nil {
			return "", errors.TokenError(
				fmt.Sprintf("Failed to read token file: %s", err.Error()),
				tokenFile,
				err,
			)
		}
		token := strings.TrimSpace(string(data))
		if len(token) == 0 {
			return "", errors.TokenError(
				"Token file is empty",
				tokenFile,
				nil,
			)
		}
		return token, nil
	}

	return "", errors.TokenError(
		fmt.Sprintf("No token provided - neither %s environment variable nor token file specified", environmentVariable),
		tokenFile,
		nil,
	)
}

func NewClient(tokenFile string) (*Client, error) {
	token, err := GetToken(tokenFile)
	if err != nil {
		return nil, err
	}

	secretsAPI, err := retryOperation(defaultInitAttempts, "Initializing 1Password client", func() (sdkSecretsAPI, error) {
		return newSDKSecrets(context.Background(), token)
	})
	if err != nil {
		return nil, errors.OnePasswordError(
			"Initializing 1Password client",
			"Failed to create 1Password SDK client - check token validity",
			err,
		)
	}

	return &Client{secrets: secretsAPI}, nil
}

// NewConnectClient creates a client that retrieves secrets from a 1Password
// Connect server. The token comes from OP_CONNECT_TOKEN or tokenFile.
func NewConnectClient(host, tokenFile string) (*Client, error) {
	if strings.TrimSpace(host) == "" {
		return nil, fmt.Errorf("1Password Connect host must not be empty")
	}

	token, err := getToken(tokenFile, "OP_CONNECT_TOKEN")
	if err != nil {
		return nil, err
	}

	return &Client{connectSecrets: newConnectSecrets(host, token)}, nil
}

func (c *Client) ResolveSecret(reference string) (string, error) {
	secrets, err := c.ResolveSecrets([]string{reference})
	if err != nil {
		return "", err
	}
	return secrets[reference], nil
}

func (c *Client) ResolveSecrets(references []string) (map[string]string, error) {
	if c.connectSecrets != nil {
		return c.resolveConnectSecrets(references)
	}

	response, err := retryOperation(defaultResolveAttempts, "Resolving 1Password references", func() (onepassword.ResolveAllResponse, error) {
		return c.secrets.ResolveAll(context.Background(), references)
	}, shouldRetryProviderError)
	if err != nil {
		return nil, errors.OnePasswordError(
			"Resolving 1Password secrets",
			"Failed to resolve 1Password references",
			classifyTopLevelProviderError(err),
		)
	}

	resolved := make(map[string]string, len(references))
	var failures []*errors.ProviderError
	for _, reference := range references {
		individual, ok := response.IndividualResponses[reference]
		if !ok {
			failures = append(failures, &errors.ProviderError{
				Kind:      errors.ProviderErrorOther,
				Reference: reference,
				Issue:     "1Password did not return a result for this reference",
			})
			continue
		}

		if individual.Error != nil {
			failures = append(failures, classifyResolveReferenceError(reference, *individual.Error))
			continue
		}
		if individual.Content == nil {
			failures = append(failures, &errors.ProviderError{
				Kind:      errors.ProviderErrorOther,
				Reference: reference,
				Issue:     "1Password returned an empty result for this reference",
			})
			continue
		}

		resolved[reference] = individual.Content.Secret
	}

	if len(failures) > 0 {
		return nil, &errors.ProviderResolutionError{Failures: failures}
	}

	return resolved, nil
}

func (c *Client) resolveConnectSecrets(references []string) (map[string]string, error) {
	resolved := make(map[string]string, len(references))
	var failures []*errors.ProviderError

	for _, reference := range references {
		value, providerError := c.resolveConnectSecret(reference)
		if providerError != nil {
			failures = append(failures, providerError)
			continue
		}
		resolved[reference] = value
	}

	if len(failures) > 0 {
		return nil, &errors.ProviderResolutionError{Failures: failures}
	}

	return resolved, nil
}

func (c *Client) resolveConnectSecret(reference string) (string, *errors.ProviderError) {
	vault, item, section, field, err := parseConnectReference(reference)
	if err != nil {
		return "", &errors.ProviderError{
			Kind:      errors.ProviderErrorInvalidReference,
			Reference: reference,
			Issue:     err.Error(),
			Cause:     err,
		}
	}

	resolvedItem, err := retryOperation(defaultResolveAttempts, "Resolving 1Password Connect reference", func() (*connectonepassword.Item, error) {
		return c.connectSecrets.GetItem(item, vault)
	})
	if err != nil {
		return "", &errors.ProviderError{
			Kind:      errors.ProviderErrorTransient,
			Reference: reference,
			Issue:     err.Error(),
			Cause:     err,
		}
	}

	for _, candidate := range resolvedItem.Fields {
		if candidate.Label != field || !matchesSection(candidate.Section, section) {
			continue
		}
		return candidate.Value, nil
	}

	return "", &errors.ProviderError{
		Kind:      errors.ProviderErrorMissingReference,
		Reference: reference,
		Issue:     "field was not found in the 1Password Connect item",
	}
}

func parseConnectReference(reference string) (vault, item, section, field string, err error) {
	if !strings.HasPrefix(reference, "op://") {
		return "", "", "", "", fmt.Errorf("reference must start with op://")
	}

	segments := strings.Split(strings.TrimPrefix(reference, "op://"), "/")
	if len(segments) != 3 && len(segments) != 4 {
		return "", "", "", "", fmt.Errorf("reference must have vault, item, and field, with an optional section")
	}
	for index, segment := range segments {
		if segment == "" {
			return "", "", "", "", fmt.Errorf("reference segment %d must not be empty", index+1)
		}
		segments[index], err = url.PathUnescape(segment)
		if err != nil {
			return "", "", "", "", fmt.Errorf("reference segment %d is not valid URL encoding: %w", index+1, err)
		}
	}

	if len(segments) == 3 {
		return segments[0], segments[1], "", segments[2], nil
	}
	return segments[0], segments[1], segments[2], segments[3], nil
}

func matchesSection(candidate *connectonepassword.ItemSection, expected string) bool {
	if expected == "" {
		// Connect represents fields added without an explicit section as part of
		// its implicit "add more" section. References without a section should
		// resolve those fields in the same way as fields with no section.
		return candidate == nil || (candidate.ID == "add more" && candidate.Label == "")
	}
	return candidate != nil && candidate.Label == expected
}

func retryOperation[T any](attempts int, operation string, fn func() (T, error), shouldRetry ...func(error) bool) (T, error) {
	var zero T
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := fn()
		if err == nil {
			return value, nil
		}

		lastErr = err
		if len(shouldRetry) > 0 && !shouldRetry[0](err) {
			break
		}
		if attempt == attempts {
			break
		}

		fmt.Fprintf(os.Stderr, "WARNING: %s attempt %d/%d failed: %v\n", operation, attempt, attempts, err)
		retrySleep(time.Duration(attempt) * time.Second)
	}

	return zero, fmt.Errorf("%s failed after %d attempts: %w", operation, attempts, lastErr)
}

func shouldRetryProviderError(err error) bool {
	var rateLimited *onepassword.RateLimitExceededError
	return !stderrors.As(err, &rateLimited)
}

func classifyTopLevelProviderError(err error) error {
	var rateLimited *onepassword.RateLimitExceededError
	if stderrors.As(err, &rateLimited) {
		return &errors.ProviderResolutionError{Failures: []*errors.ProviderError{{
			Kind:  errors.ProviderErrorRateLimited,
			Issue: "1Password rate limit exceeded; wait for the provider reset window before retrying",
			Cause: err,
		}}}
	}
	return &errors.ProviderResolutionError{Failures: []*errors.ProviderError{{
		Kind:  errors.ProviderErrorTransient,
		Issue: err.Error(),
		Cause: err,
	}}}
}

func classifyResolveReferenceError(reference string, err onepassword.ResolveReferenceError) *errors.ProviderError {
	switch err.Type {
	case onepassword.ResolveReferenceErrorTypeVariantVaultNotFound,
		onepassword.ResolveReferenceErrorTypeVariantItemNotFound,
		onepassword.ResolveReferenceErrorTypeVariantFieldNotFound,
		onepassword.ResolveReferenceErrorTypeVariantNoMatchingSections:
		return &errors.ProviderError{Kind: errors.ProviderErrorMissingReference, Reference: reference, Issue: string(err.Type)}
	case onepassword.ResolveReferenceErrorTypeVariantParsing:
		return &errors.ProviderError{Kind: errors.ProviderErrorInvalidReference, Reference: reference, Issue: string(err.Parsing())}
	case onepassword.ResolveReferenceErrorTypeVariantTooManyVaults,
		onepassword.ResolveReferenceErrorTypeVariantTooManyItems,
		onepassword.ResolveReferenceErrorTypeVariantTooManyMatchingFields:
		return &errors.ProviderError{Kind: errors.ProviderErrorAmbiguousReference, Reference: reference, Issue: string(err.Type)}
	default:
		return &errors.ProviderError{Kind: errors.ProviderErrorOther, Reference: reference, Issue: string(err.Type)}
	}
}
