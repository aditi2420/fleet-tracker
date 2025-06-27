package model

import (
	"time"

	"encoding/json"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Status is the JSON blob we cache & store.
type Status struct {
	Location  [2]float64 `json:"location"` // [long, lat]
	Speed     float64    `json:"speed"`
	Timestamp time.Time  `json:"timestamp"`
}

// Vehicle maps to the "vehicle" table.
// The struct tags satisfy both JSON (HTTP responses) and GORM.
type Vehicle struct {
	ID          uuid.UUID      `json:"id"           gorm:"type:uuid;primaryKey"`
	PlateNumber string         `json:"plate_number" gorm:"uniqueIndex"`
	LastStatus  datatypes.JSON `json:"last_status"`
}

func (v *Vehicle) DecodeStatus() (Status, error) {
	var s Status
	err := json.Unmarshal(v.LastStatus, &s)
	return s, err
}

type InputRequestPayload struct {
	VehicleID   uuid.UUID `json:"vehicle_id"`
	PlateNumber string    `json:"plate_number"`
	Status      Status    `json:"status"`
}
