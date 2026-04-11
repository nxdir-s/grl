package domain

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

const (
	DefaultActiveEnvID     string = ""
	DefaultMethod          string = "GET"
	DefaultTimeoutSeconds  int    = 30
	DefaultFollowRedirects bool   = true
	DefaultHistoryLimit    int    = 100
)

type Configs struct {
	service ports.ConfigService
}

func NewConfigs(service ports.ConfigService) *Configs {
	return &Configs{
		service: service,
	}
}

func (d *Configs) Get(ctx context.Context) (*valobj.Config, error) {
	config, err := d.service.Get(ctx)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return d.defaultConfig(), nil
	}

	defaults := d.defaultConfig()

	if len(config.DefaultMethod) == 0 {
		config.DefaultMethod = defaults.DefaultMethod
	}

	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = defaults.TimeoutSeconds
	}

	if config.HistoryLimit == 0 {
		config.HistoryLimit = defaults.HistoryLimit
	}

	return config, nil
}

func (d *Configs) Save(ctx context.Context, cfg *valobj.Config) error {
	return d.service.Save(ctx, cfg)
}

func (d *Configs) Defaults() *valobj.Config {
	return d.defaultConfig()
}

func (d *Configs) defaultConfig() *valobj.Config {
	return &valobj.Config{
		ActiveEnvID:     DefaultActiveEnvID,
		DefaultMethod:   DefaultMethod,
		TimeoutSeconds:  DefaultTimeoutSeconds,
		FollowRedirects: DefaultFollowRedirects,
		HistoryLimit:    DefaultHistoryLimit,
	}
}
