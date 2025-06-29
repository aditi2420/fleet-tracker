package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVehicleRepo_Get(t *testing.T) {
	db := setupTestDB(t)
	tripRepo := NewTripRepo(db)
	repo := NewVehicleRepo(db, tripRepo)

	tests := []struct {
		name        string
		setupData   func() uuid.UUID
		expectedErr bool
	}{
		{
			name: "positive - vehicle exists",
			setupData: func() uuid.UUID {
				vehicleID := uuid.New()
				status := model.Status{
					Location:  [2]float64{123.456, 789.012},
					Speed:     60.5,
					Timestamp: time.Now(),
				}
				statusJSON, _ := json.Marshal(status)
				vehicle := model.Vehicle{
					ID:          vehicleID,
					PlateNumber: "ABC123",
					LastStatus:  statusJSON,
				}
				db.Create(&vehicle)
				return vehicleID
			},
			expectedErr: false,
		},
		{
			name: "negative - vehicle not found",
			setupData: func() uuid.UUID {
				return uuid.New() // we are not creating entry so will not be present.
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicleID := tt.setupData()

			result, err := repo.Get(context.Background(), vehicleID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, vehicleID, result.ID)
				assert.Equal(t, "ABC123", result.PlateNumber)
			}
		})
	}
}

func TestVehicleRepo_UpsertStatusAndInsertTrip(t *testing.T) {
	db := setupTestDB(t)
	tripRepo := NewTripRepo(db)
	repo := NewVehicleRepo(db, tripRepo)

	tests := []struct {
		name        string
		setupData   func() (uuid.UUID, string, model.Status, model.Trips)
		expectedErr bool
	}{
		{
			name: "positive - insert vehicle and trip",
			setupData: func() (uuid.UUID, string, model.Status, model.Trips) {
				vehicleID := uuid.New()
				plate := "GHI789"
				status := model.Status{
					Location:  [2]float64{555.666, 777.888},
					Speed:     80.0,
					Timestamp: time.Now(),
				}
				trip := model.Trips{
					ID:        uuid.New(),
					VehicleID: vehicleID,
					StartTime: time.Now(),
					AvgSpeed:  80.0,
				}
				return vehicleID, plate, status, trip
			},
			expectedErr: false,
		},
		{
			name: "positive - update vehicle and insert trip",
			setupData: func() (uuid.UUID, string, model.Status, model.Trips) {
				vehicleID := uuid.New()
				plate := "JKL012"
				status := model.Status{
					Location:  [2]float64{111.222, 333.444},
					Speed:     65.0,
					Timestamp: time.Now(),
				}
				trip := model.Trips{
					ID:        uuid.New(),
					VehicleID: vehicleID,
					StartTime: time.Now(),
					AvgSpeed:  65.0,
				}

				// Create initial vehicle
				initialStatus := model.Status{
					Location:  [2]float64{999.888, 777.666},
					Speed:     30.0,
					Timestamp: time.Now().Add(-1 * time.Hour),
				}
				initialStatusJSON, _ := json.Marshal(initialStatus)
				vehicle := model.Vehicle{
					ID:          vehicleID,
					PlateNumber: plate,
					LastStatus:  initialStatusJSON,
				}
				db.Create(&vehicle)

				return vehicleID, plate, status, trip
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicleID, plate, status, trip := tt.setupData()

			err := repo.UpsertStatusAndInsertTrip(context.Background(), vehicleID, plate, status, trip)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify vehicle was updated
				var vehicle model.Vehicle
				err := db.First(&vehicle, "id = ?", vehicleID).Error
				assert.NoError(t, err)
				assert.Equal(t, vehicleID, vehicle.ID)
				assert.Equal(t, plate, vehicle.PlateNumber)

				savedStatus, err := vehicle.DecodeStatus()
				assert.NoError(t, err)
				assert.Equal(t, status.Location, savedStatus.Location)
				assert.Equal(t, status.Speed, savedStatus.Speed)

				// Verify trip was inserted
				var savedTrip model.Trips
				err = db.First(&savedTrip, "id = ?", trip.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, trip.ID, savedTrip.ID)
				assert.Equal(t, trip.VehicleID, savedTrip.VehicleID)
				assert.Equal(t, trip.AvgSpeed, savedTrip.AvgSpeed)
			}
		})
	}
}
