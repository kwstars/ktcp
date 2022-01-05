package storage

import (
	"context"
)

type Saver interface {
	Save(ctx context.Context) error
}
