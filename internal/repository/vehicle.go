package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aditi2420/fleet-tracker/internal/model"
)

type VehicleRepo struct {
	db       *gorm.DB
	tripRepo *TripRepo // needed for the composite Tx
}

func NewVehicleRepo(db *gorm.DB, tripRepo *TripRepo) *VehicleRepo {
	return &VehicleRepo{db: db, tripRepo: tripRepo}
}

// Get returns the current vehicle
func (r *VehicleRepo) Get(ctx context.Context, id uuid.UUID) (model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error
	return v, err
}

// UpsertStatus updates status of a vehicle
func (r *VehicleRepo) UpsertStatus(
	ctx context.Context,
	id uuid.UUID, plate string,
	st model.Status,
	tx *gorm.DB,
) error {
	b, _ := json.Marshal(st)
	return tx.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&model.Vehicle{
			ID:          id,
			PlateNumber: plate,
			LastStatus:  datatypes.JSON(b),
		}).Error
}

// UpsertStatusAndInsertTrip insert data in both trip and vehicle status(used in ingest)
func (r *VehicleRepo) UpsertStatusAndInsertTrip(
	ctx context.Context,
	id uuid.UUID, plate string,
	st model.Status,
	trip model.Trips,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.UpsertStatus(ctx, id, plate, st, tx); err != nil {
			return err
		}
		if err := r.tripRepo.Create(ctx, trip, tx); err != nil {
			return err
		}
		return nil
	})
}
