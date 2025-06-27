package main

import (
	"github.com/aditi2420/fleet-tracker/internal/middleware"
	"github.com/aditi2420/fleet-tracker/internal/stream"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/golang-jwt/jwt/v5"
	//"github.com/google/uuid"

	"context"
	"github.com/aditi2420/fleet-tracker/internal/cache"
	"github.com/aditi2420/fleet-tracker/internal/controller"
	"github.com/aditi2420/fleet-tracker/internal/repository"
	"github.com/aditi2420/fleet-tracker/internal/service"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	//env variable read
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		log.Fatal("PG_DSN env var is missing")
	}

	redisAddr := os.Getenv("REDIS_ADDR") // redis:6379

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//check db connection and redis
	db, err := repository.New(dsn) // GORM *gorm.DB
	if err != nil {
		log.Fatalf("gorm: %v", err)
	}
	redisCache, err := cache.NewRedis(redisAddr, "", 0, 5*time.Minute)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// JWT middleware
	secret := []byte(os.Getenv("JWT_SIGN_KEY"))
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.NewJWT(secret))
	r.Use(middleware.HTTPLogger())

	//add routes
	tripRepo := repository.NewTripRepo(db)
	vehicleRepo := repository.NewVehicleRepo(db, tripRepo)

	svc := service.New(vehicleRepo, tripRepo, redisCache)

	// API routes
	api := r.Group("/api/vehicle")
	{
		api.GET("/status", controller.GetStatusHandler(svc))
		api.GET("/trips", controller.GetTripsHandler(svc))
		api.POST("/ingest", controller.IngestHandler(svc))
	}

	//start the producer and consumer for every 2 min streaming
	vehID := uuid.New()
	ch := make(chan stream.Payload, 30)
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		slog.Info("shutdown signal received")
		cancel()
	}()

	go stream.Produce(ctx, ch, vehID)
	go stream.Consumer(ctx, ch, svc)

	log.Printf("⇢ listening on :%s …", port)
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

}
