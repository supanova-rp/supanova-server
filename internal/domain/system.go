package domain

import (
	"context"
)

type SystemRepository interface {
	PingDB(context.Context) error
}