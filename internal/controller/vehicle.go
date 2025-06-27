package controller

import (
	"net/http"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/aditi2420/fleet-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VehicleController struct {
	VehicleService service.VehicleService
}

func GetStatusHandler1(svc service.VehicleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Query("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad uuid"})
			return
		}
		st, err := svc.CurrentStatus(c, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, st)
	}
}

func GetStatusHandler(svc service.VehicleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Query("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad uuid"})
			return
		}
		user := c.GetString("user")
		st, err := svc.CurrentStatus(c, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": st, "user": user})
	}
}

func GetTripsHandler(svc service.VehicleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Query("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad uuid"})
			return
		}
		tr, err := svc.ListTrips(c, id, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tr)
	}
}

func IngestHandler(svc service.VehicleService) gin.HandlerFunc {
	type payload struct {
		VehicleID uuid.UUID    `json:"vehicle_id"`
		Status    model.Status `json:"status"`
	}
	return func(c *gin.Context) {
		var p model.InputRequestPayload
		if err := c.BindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := svc.Ingest(c, p.VehicleID, p.PlateNumber, p.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	}
}
