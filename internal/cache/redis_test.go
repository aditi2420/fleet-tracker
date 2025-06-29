package cache

import (
	"context"
	"testing"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedis(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		password    string
		db          int
		ttl         time.Duration
		expectError bool
	}{
		{
			name:        "positive - valid configuration",
			addr:        "localhost:6379",
			password:    "",
			db:          0,
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "error - invalid address",
			addr:        "invalid:address",
			password:    "",
			db:          0,
			ttl:         5 * time.Minute,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewRedis(tt.addr, tt.password, tt.db, tt.ttl)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)

				if cache != nil {
					defer cache.Close()

					if tt.ttl == 0 {
						assert.Equal(t, defaultTTL, cache.TTL())
					} else {
						assert.Equal(t, tt.ttl, cache.TTL())
					}
				}
			}
		})
	}
}

func TestRedisCache_GetStatus(t *testing.T) {

	addr := "localhost:6379"

	cache, err := NewRedis(addr, "", 0, 5*time.Minute)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	tests := []struct {
		name        string
		setupData   func() (uuid.UUID, *model.Status)
		expected    *model.Status
		expectError bool
	}{
		{
			name: "success - status exists",
			setupData: func() (uuid.UUID, *model.Status) {
				vehicleID := uuid.New()
				status := &model.Status{
					Location:  [2]float64{123.456, 789.012},
					Speed:     60.5,
					Timestamp: time.Now(),
				}
				err := cache.SetStatus(ctx, vehicleID, *status)
				require.NoError(t, err)
				return vehicleID, status
			},
			expected: &model.Status{
				Location:  [2]float64{123.456, 789.012},
				Speed:     60.5,
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name: "success - status not found",
			setupData: func() (uuid.UUID, *model.Status) {
				vehicleID := uuid.New()
				return vehicleID, nil
			},
			expected:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicleID, _ := tt.setupData()

			result, err := cache.GetStatus(ctx, vehicleID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != nil {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.Location, result.Location)
					assert.Equal(t, tt.expected.Speed, result.Speed)
				} else {
					assert.Nil(t, result)
				}
			}
		})
	}
}

func TestRedisCache_SetStatus(t *testing.T) {
	addr := "localhost:6379"

	cache, err := NewRedis(addr, "", 0, 5*time.Minute)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	vehicleID := uuid.New()
	tests := []struct {
		name        string
		vehicleID   uuid.UUID
		status      model.Status
		expectError bool
	}{
		{
			name:      "positive - set status",
			vehicleID: vehicleID,
			status: model.Status{
				Location:  [2]float64{123.456, 789.012},
				Speed:     60.5,
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name:      "success - update existing status",
			vehicleID: vehicleID,
			status: model.Status{
				Location:  [2]float64{456.789, 123.456},
				Speed:     75.0,
				Timestamp: time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.SetStatus(ctx, tt.vehicleID, tt.status)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the status was set correctly
				result, err := cache.GetStatus(ctx, tt.vehicleID)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.status.Location, result.Location)
				assert.Equal(t, tt.status.Speed, result.Speed)
			}
		})
	}
}
