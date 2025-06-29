package repository

import (
	"context"
	"testing"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate the models
	err = db.AutoMigrate(&model.Vehicle{}, &model.Trips{})
	require.NoError(t, err)

	return db
}

func TestTripRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTripRepo(db)

	tests := []struct {
		name        string
		setupData   func() (model.Trips, *gorm.DB)
		expectedErr bool
	}{
		{
			name: "positive - create new trip",
			setupData: func() (model.Trips, *gorm.DB) {
				trip := model.Trips{
					ID:        uuid.New(),
					VehicleID: uuid.New(),
					StartTime: time.Now(),
					AvgSpeed:  65.0,
					Mileage:   100.5,
				}
				return trip, db
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trip, tx := tt.setupData()

			err := repo.Create(context.Background(), trip, tx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify the trip was created
				var savedTrip model.Trips
				err := db.First(&savedTrip, "id = ?", trip.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, trip.ID, savedTrip.ID)
				assert.Equal(t, trip.VehicleID, savedTrip.VehicleID)
				assert.Equal(t, trip.AvgSpeed, savedTrip.AvgSpeed)
				assert.Equal(t, trip.Mileage, savedTrip.Mileage)
			}
		})
	}
}

func TestTripRepo_ListRecent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTripRepo(db)

	tests := []struct {
		name        string
		setupData   func() (uuid.UUID, time.Duration, []model.Trips)
		expectedLen int
		expectedErr bool
	}{
		{
			name: "positive - returns recent trips",
			setupData: func() (uuid.UUID, time.Duration, []model.Trips) {
				vehicleID := uuid.New()
				since := 2 * time.Hour
				
				// Create trips with different timestamps
				now := time.Now()
				trips := []model.Trips{
					{
						ID:        uuid.New(),
						VehicleID: vehicleID,
						StartTime: now.Add(-30 * time.Minute), // Recent
						AvgSpeed:  65.0,
					},
					{
						ID:        uuid.New(),
						VehicleID: vehicleID,
						StartTime: now.Add(-1 * time.Hour), // Recent
						AvgSpeed:  70.0,
					},
					{
						ID:        uuid.New(),
						VehicleID: vehicleID,
						StartTime: now.Add(-3 * time.Hour), // Too old
						AvgSpeed:  55.0,
					},
				}
				
				// Insert trips into database
				for _, trip := range trips {
					db.Create(&trip)
				}
				
				return vehicleID, since, trips[:2]
			},
			expectedLen: 2,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicleID, since, _ := tt.setupData()

			result, err := repo.ListRecent(context.Background(), vehicleID, since)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)

				if len(result) > 1 {
					assert.True(t, result[0].StartTime.After(result[1].StartTime))
				}

				cutoffTime := time.Now().Add(-since)
				for _, trip := range result {
					assert.Equal(t, vehicleID, trip.VehicleID)
					assert.True(t, trip.StartTime.After(cutoffTime))
				}
			}
		})
	}
}

func TestTripRepo_Create_WithTransaction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTripRepo(db)

	trip := model.Trips{
		ID:        uuid.New(),
		VehicleID: uuid.New(),
		StartTime: time.Now(),
		AvgSpeed:  65.0,
	}

	// Test with transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		return repo.Create(context.Background(), trip, tx)
	})

	assert.NoError(t, err)
	
	// Verify the trip was created
	var savedTrip model.Trips
	err = db.First(&savedTrip, "id = ?", trip.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, trip.ID, savedTrip.ID)
}
