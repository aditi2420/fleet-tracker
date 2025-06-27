package stream

import (
	"context"
	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/google/uuid"
	"log/slog"
	"math/rand"
	"time"
)

// Produce generate mock data
func Produce(ctx context.Context, out chan<- Payload, id uuid.UUID) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	rand.Seed(time.Now().UnixNano())
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			slog.Info("Producing payload for vehicle %s", id.String())
			out <- Payload{ //writing into the channel:out
				VehicleID: id,
				Plate:     id.String(),
				Status:    getMockedStatus(t),
			}
		}
	}
}

func getMockedStatus(now time.Time) model.Status {
	const lat, long = 25.1972, 55.2744
	return model.Status{
		Location: [2]float64{
			lat + rand.Float64()*0.02,
			long + rand.Float64()*0.02,
		},
		Speed:     40 + rand.Float64(),
		Timestamp: now.UTC(),
	}
}
