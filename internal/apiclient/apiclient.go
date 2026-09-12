// Package apiclient resolves a qrocodile-api-go Client from the environment and the local
// config file.
package apiclient

import (
	"fmt"
	"os"

	qrocodile "github.com/qrocodile-io/qrocodile-api-go"

	"github.com/qrocodile-io/qrocodile-cli/internal/config"
)

// EnvAPIKey always overrides the stored key — the load-bearing mechanism for CI use, matching
// GH_TOKEN/AWS_ACCESS_KEY_ID conventions.
const EnvAPIKey = "QROCODILE_API_KEY"

// EnvBaseURL overrides the API base URL, e.g. to target a dev/staging instance.
const EnvBaseURL = "QROCODILE_API_BASE_URL"

// Resolve builds an authenticated client: EnvAPIKey if set, otherwise the key stored by
// `qrocodile auth login`. Returns an error naming both options if neither is available.
func Resolve() (*qrocodile.Client, error) {
	key := os.Getenv(EnvAPIKey)
	if key == "" {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		key = cfg.APIKey()
	}
	if key == "" {
		return nil, fmt.Errorf("not logged in — run `qrocodile auth login`, or set %s", EnvAPIKey)
	}
	return newClient(key), nil
}

// Unauthenticated builds a client with no API key, for the key-registration flow
// (RegisterKey/ConfirmKey), which needs none.
func Unauthenticated() *qrocodile.Client {
	return newClient("")
}

func newClient(apiKey string) *qrocodile.Client {
	var opts []qrocodile.ClientOption
	if baseURL := os.Getenv(EnvBaseURL); baseURL != "" {
		opts = append(opts, qrocodile.WithBaseURL(baseURL))
	}
	return qrocodile.NewClient(apiKey, opts...)
}
