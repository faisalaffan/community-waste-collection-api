package worker

import (
	"context"
	"log"
	"time"

	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type OrganicCancelWorker struct {
	pickupRepo repository.PickupRepository
	interval   time.Duration
	olderThan  time.Duration
}

func NewOrganicCancelWorker(repo repository.PickupRepository) *OrganicCancelWorker {
	return &OrganicCancelWorker{
		pickupRepo: repo,
		interval:   1 * time.Hour,
		olderThan:  3 * 24 * time.Hour,
	}
}

func (w *OrganicCancelWorker) Start(ctx context.Context) {
	log.Println("[worker] organic cancel worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[worker] organic cancel worker stopped")
			return
		case <-ticker.C:
			affected, err := w.pickupRepo.CancelOrganicPending(w.olderThan)
			if err != nil {
				log.Printf("[worker] error canceling organic pickups: %v", err)
			} else if affected > 0 {
				log.Printf("[worker] canceled %d organic pickups", affected)
			}
		}
	}
}
