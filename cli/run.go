package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dtonair/asana-cli/asana"
	"github.com/dtonair/asana-cli/config"
)

// pageSize is the per-request page size used for paginated endpoints, matching
// the extension's behavior.
const pageSize = 50

// maxPages bounds pagination, matching the extension.
const maxPages = 10

// buildClient loads config and constructs an Asana client honoring the
// persistent flags. Config failures are usage errors (exit code 2).
//
// ASANA_API_BASE overrides the API base URL; it exists only to point tests at
// an httptest server and is not a documented user-facing flag.
func buildClient() (*asana.Client, config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, config.Config{}, &usageError{err: err}
	}

	httpClient := &http.Client{Timeout: opts.timeout}
	var copts []asana.Option
	if base := strings.TrimSpace(os.Getenv("ASANA_API_BASE")); base != "" {
		copts = append(copts, asana.WithBaseURL(base))
	}
	if opts.verbose {
		copts = append(copts, asana.WithVerbose(os.Stderr))
	}
	return asana.NewClient(cfg.AccessToken, httpClient, copts...), cfg, nil
}

// withTimeout derives a context bounded by the --timeout flag.
func withTimeout(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return context.WithTimeout(cmd.Context(), opts.timeout)
}

// validateLimit enforces the 1..100 bound (usage error on violation).
func validateLimit(limit int) error {
	if limit < 1 || limit > 100 {
		return usageErrorf("--limit must be between 1 and 100, got %d", limit)
	}
	return nil
}

// requireFlag returns a trimmed required string value or a usage error.
func requireFlag(name, value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", usageErrorf("--%s is required", name)
	}
	return v, nil
}

// query helpers

func appendOptFields(q url.Values, optFields string) {
	if v := strings.TrimSpace(optFields); v != "" {
		q.Set("opt_fields", v)
	}
}

func querySuffix(q url.Values) string {
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

// requestData performs a request and unwraps the top-level "data" field.
func requestData(ctx context.Context, c *asana.Client, method, path string, body any) (json.RawMessage, error) {
	raw, err := c.Request(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return env.Data, nil
}
