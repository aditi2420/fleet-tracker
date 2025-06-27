package service

import (
	"context"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/cache"
	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/aditi2420/fleet-tracker/internal/repository"
	"github.com/google/uuid"
)

// VehicleService is consumed by HTTP handlers (and the stream processor).
type VehicleService interface {
	CurrentStatus(ctx context.Context, vehicleID uuid.UUID) (model.Status, error)
	ListTrips(ctx context.Context, vehicleID uuid.UUID, since time.Duration) ([]model.Trips, error)
	Ingest(ctx context.Context, vehicleID uuid.UUID, plate string, s model.Status) error
}

type service struct {
	vehRepo  *repository.VehicleRepo
	tripRepo *repository.TripRepo
	cache    cache.VehicleCache
}

func New(v *repository.VehicleRepo, t *repository.TripRepo, c cache.VehicleCache) VehicleService {
	return &service{vehRepo: v, tripRepo: t, cache: c}
}

func (s *service) CurrentStatus(ctx context.Context, id uuid.UUID) (model.Status, error) {
	//check in cache
	if st, _ := s.cache.GetStatus(ctx, id); st != nil {
		return *st, nil
	}
	veh, err := s.vehRepo.Get(ctx, id)
	if err != nil {
		return model.Status{}, err
	}
	st, err := veh.DecodeStatus()
	if err != nil {
		return model.Status{}, err
	}

	_ = s.cache.SetStatus(ctx, id, st) // fire‑and‑forget
	return st, nil
}

func (s *service) ListTrips(ctx context.Context, id uuid.UUID, since time.Duration) ([]model.Trips, error) {
	return s.tripRepo.ListRecent(ctx, id, since)
}

func (s *service) Ingest(ctx context.Context, id uuid.UUID, plate string, st model.Status) error {
	trip := model.Trips{
		ID:        uuid.New(),
		VehicleID: id,
		StartTime: st.Timestamp,
		AvgSpeed:  st.Speed,
	}

	if err := s.vehRepo.UpsertStatusAndInsertTrip(ctx, id, plate, st, trip); err != nil {
		return err
	}

	return s.cache.SetStatus(ctx, id, st)
}
