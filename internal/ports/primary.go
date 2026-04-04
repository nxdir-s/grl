package ports

import "context"

type CLI interface {
	RunTUI(ctx context.Context) error
}
