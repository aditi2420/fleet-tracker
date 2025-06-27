package cache

import (
	"context"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/google/uuid"
)

// VehicleCache encapsulates just the behaviour we need today.
// Add more methods later (e.g. trip aggregates) without leaking go‑redis.
type VehicleCache interface {
	GetStatus(ctx context.Context, id uuid.UUID) (*model.Status, error)
	SetStatus(ctx context.Context, id uuid.UUID, s model.Status) error
	TTL() time.Duration
	Close() error
}
