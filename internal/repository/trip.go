package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aditi2420/fleet-tracker/internal/model"
)

type TripRepo struct {
	db *gorm.DB
}

func NewTripRepo(db *gorm.DB) *TripRepo {
	return &TripRepo{db}
}

// Create adds new location/trip
func (r *TripRepo) Create(ctx context.Context, t model.Trips, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(&t).Error
}

// ListRecent fetches all recent trips after given duration
func (r *TripRepo) ListRecent(
	ctx context.Context,
	id uuid.UUID,
	since time.Duration,
) ([]model.Trips, error) {
	fmt.Println("valuesss are", id, since)
	var res []model.Trips
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND start_time >= ?", id, time.Now().Add(-since)).
		Order("start_time DESC").
		Find(&res).Error
	return res, err
}
