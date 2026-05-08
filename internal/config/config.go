package config

import (
	"context"
	"os"
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
)

type Option func(c *Config) error

func WithCredentials() Option {
	return func(c *Config) error {
		clientId := os.Getenv(ClientID)
		secret := os.Getenv(ClientSecret)
		oauth := os.Getenv(OAuthURL)

		if len(clientId) != 0 && len(secret) != 0 && len(oauth) != 0 {
			c.Credentials = true
		}

		c.ClientId = clientId
		c.ClientSecret = secret
		c.OAuthURL = oauth

		return nil
	}
}

type Config struct {
	ClientId     string
	ClientSecret string
	OAuthURL     string
	Credentials  bool
}

func New(ctx context.Context, opts ...Option) (*Config, error) {
	config := &Config{}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}
