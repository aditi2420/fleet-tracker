package model

import (
	"github.com/google/uuid"
	"time"
)

// Trips maps to "trips" table.
type Trips struct {
	ID        uuid.UUID  `json:"id"         gorm:"type:uuid;primaryKey"`
	VehicleID uuid.UUID  `json:"vehicle_id" gorm:"type:uuid;index"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Mileage   float64    `json:"mileage"`
	AvgSpeed  float64    `json:"avg_speed"`
}
