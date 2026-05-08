package config

import (
	"context"
	"os"
	"strings"
)

type ErrEnvVar struct {
	evar string
}

func (e *ErrEnvVar) Error() string {
	return e.evar + " not found in environment"
}

const (
	ClientID     string = "CLIENT_ID"
	ClientSecret string = "CLIENT_SECRET"
	OAuthURL     string = "OAUTH_URL"
	OAuthScopes  string = "OAUTH_SCOPES"

	Delimiter string = ","
)

type Option func(c *Config) error

func WithCredentials() Option {
	return func(c *Config) error {
		clientId := os.Getenv(ClientID)
		secret := os.Getenv(ClientSecret)
		oauth := os.Getenv(OAuthURL)
		scopes := os.Getenv(OAuthScopes)

		if len(clientId) != 0 && len(secret) != 0 && len(oauth) != 0 {
			c.Credentials = true
		}

		authScopes := make([]string, 0)
		if len(scopes) > 0 {
			authScopes = append(authScopes, strings.Split(scopes, Delimiter)...)
		}

		c.ClientId = clientId
		c.ClientSecret = secret
		c.OAuthURL = oauth
		c.OAuthScopes = append(c.OAuthScopes, authScopes...)

		return nil
	}
}

type Config struct {
	ClientId     string
	ClientSecret string
	OAuthURL     string
	OAuthScopes  []string
	Credentials  bool
}

func New(ctx context.Context, opts ...Option) (*Config, error) {
	config := &Config{
		OAuthScopes: make([]string, 0),
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}
