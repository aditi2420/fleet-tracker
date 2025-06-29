package stream

import (
	"context"
	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/aditi2420/fleet-tracker/internal/service"
	"github.com/google/uuid"
	"log/slog"
)

// Consumer calls service layer internally
func Consumer(ctx context.Context, in <-chan Payload, svc service.VehicleService) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-in:
			if err := svc.Ingest(ctx, payload.VehicleID, payload.Plate, payload.Status); err != nil {
				slog.Error("ingest api failed for vehicle id", "vehicle", payload.VehicleID, "err", err)
				continue
			}

			slog.Info("successfully ingested vehicleID",
				slog.String("vehicle_id", payload.VehicleID.String()),
				slog.Float64("lat", payload.Status.Location[1]),
				slog.Float64("long", payload.Status.Location[0]),
			)
		}
	}
}

// Payload holds the channel input/output.
type Payload struct {
	VehicleID uuid.UUID
	Plate     string
	model.Status
}
