package domain

import (
	"context"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
)

type Environments struct {
	service ports.EnvironmentService
	configs ports.Configs
}

func NewEnvironments(service ports.EnvironmentService, configs ports.Configs) *Environments {
	return &Environments{
		service: service,
		configs: configs,
	}
}

func (d *Environments) Create(ctx context.Context, id string, name string) (*entity.Environment, error) {
	env := &entity.Environment{
		ID:        id,
		Name:      name,
		Variables: make(map[string]string),
	}

	if err := d.service.Save(ctx, env); err != nil {
		return nil, err
	}

	return env, nil
}

func (d *Environments) List(ctx context.Context) ([]entity.Environment, error) {
	return d.service.List(ctx)
}

func (d *Environments) Get(ctx context.Context, id string) (*entity.Environment, error) {
	return d.service.Get(ctx, id)
}

func (d *Environments) Save(ctx context.Context, env *entity.Environment) error {
	return d.service.Save(ctx, env)
}

func (d *Environments) Delete(ctx context.Context, id string) error {
	cfg, err := d.configs.Get(ctx)
	if err != nil {
		return err
	}

	if cfg.ActiveEnvID == id {
		cfg.ActiveEnvID = ""

		if err := d.configs.Save(ctx, cfg); err != nil {
			return err
		}
	}

	return d.service.Delete(ctx, id)
}

func (d *Environments) GetActive(ctx context.Context) (*entity.Environment, error) {
	cfg, err := d.configs.Get(ctx)
	if err != nil {
		return nil, err
	}

	if len(cfg.ActiveEnvID) == 0 {
		return nil, nil
	}

	env, err := d.service.Get(ctx, cfg.ActiveEnvID)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (d *Environments) SetActive(ctx context.Context, id string) error {
	cfg, err := d.configs.Get(ctx)
	if err != nil {
		return err
	}

	cfg.ActiveEnvID = id

	return d.configs.Save(ctx, cfg)
}

func (d *Environments) ActiveVars(ctx context.Context) map[string]string {
	env, err := d.GetActive(ctx)
	if err != nil {
		return map[string]string{}
	}

	if env == nil {
		return map[string]string{}
	}

	return env.Variables
}
