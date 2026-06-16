package main

import (
	"context"
	"fmt"
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
	"gorm.io/gorm"
)

type (
	databaseOpener func(dsn string) (*gorm.DB, error)
	s3Creator      func(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*storage.S3Client, error)
)

//	@title			Community Waste Collection API
//	@version		1.0
//	@description	REST API for managing community waste collection — households, pickups, payments, reports.
//	@contact.name	Muhammad Faisal Affan
//	@host			localhost:8080
//	@BasePath		/api
//	@schemes		http
func main() {
	if err := run(config.Load, database.NewPostgres, storage.NewS3); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func run(load func() (*config.Config, error), openDB databaseOpener, createS3 s3Creator) error {
	cfg, err := load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := openDB(cfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	s3Client, err := createS3(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		return fmt.Errorf("failed to connect to S3: %w", err)
	}

	householdRepo := repository.NewHouseholdRepository(db)
	pickupRepo := repository.NewPickupRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	householdSvc := service.NewHouseholdService(householdRepo)
	pickupSvc := service.NewPickupService(pickupRepo, paymentRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, s3Client)
	reportSvc := service.NewReportService(db)

	hh := handler.NewHouseholdHandler(householdSvc)
	ph := handler.NewPickupHandler(pickupSvc)
	pmh := handler.NewPaymentHandler(paymentSvc)
	rh := handler.NewReportHandler(reportSvc)

	app := router.Setup(hh, ph, pmh, rh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	organicWorker := worker.NewOrganicCancelWorker(pickupRepo)
	go organicWorker.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down...")
		cancel()
		app.Shutdown()
	}()

	log.Printf("server starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
