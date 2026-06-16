package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/faisalaffan/community-waste-collection-api/internal/config"
	"github.com/faisalaffan/community-waste-collection-api/internal/handler"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
	"github.com/faisalaffan/community-waste-collection-api/internal/router"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
	"github.com/faisalaffan/community-waste-collection-api/internal/worker"
	"github.com/faisalaffan/community-waste-collection-api/pkg/database"
	"github.com/faisalaffan/community-waste-collection-api/pkg/storage"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Database
	db, err := database.NewPostgres(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// S3 Storage
	s3Client, err := storage.NewS3(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		log.Fatalf("failed to connect to S3: %v", err)
	}

	// Repositories
	householdRepo := repository.NewHouseholdRepository(db)
	pickupRepo := repository.NewPickupRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// Services
	householdSvc := service.NewHouseholdService(householdRepo)
	pickupSvc := service.NewPickupService(pickupRepo, paymentRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, s3Client)
	reportSvc := service.NewReportService(db)

	// Handlers
	hh := handler.NewHouseholdHandler(householdSvc)
	ph := handler.NewPickupHandler(pickupSvc)
	pmh := handler.NewPaymentHandler(paymentSvc)
	rh := handler.NewReportHandler(reportSvc)

	// Router
	app := router.Setup(hh, ph, pmh, rh)

	// Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	organicWorker := worker.NewOrganicCancelWorker(pickupRepo)
	go organicWorker.Start(ctx)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down...")
		cancel()
		app.Shutdown()
	}()

	// Start server
	log.Printf("server starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
