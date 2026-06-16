package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

// mockPickupRepoWorker implements repository.PickupRepository for worker tests
type mockPickupRepoWorker struct {
	createFn               func(p *domain.WastePickup) error
	findByIDFn             func(id uuid.UUID) (*domain.WastePickup, error)
	findAllFn              func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
	updateFn               func(p *domain.WastePickup) error
	cancelOrganicPendingFn func(olderThan time.Duration) (int64, error)
}

func (m *mockPickupRepoWorker) Create(p *domain.WastePickup) error {
	if m.createFn != nil {
		return m.createFn(p)
	}
	return nil
}
func (m *mockPickupRepoWorker) FindByID(id uuid.UUID) (*domain.WastePickup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, nil
}
func (m *mockPickupRepoWorker) FindAll(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
	if m.findAllFn != nil {
		return m.findAllFn(filter)
	}
	return nil, 0, nil
}
func (m *mockPickupRepoWorker) Update(p *domain.WastePickup) error {
	if m.updateFn != nil {
		return m.updateFn(p)
	}
	return nil
}
func (m *mockPickupRepoWorker) CancelOrganicPending(olderThan time.Duration) (int64, error) {
	if m.cancelOrganicPendingFn != nil {
		return m.cancelOrganicPendingFn(olderThan)
	}
	return 0, nil
}

func TestNewOrganicCancelWorker(t *testing.T) {
	w := NewOrganicCancelWorker(nil)
	assert.NotNil(t, w)
}

func TestOrganicCancelWorker_Start_ContextCancel(t *testing.T) {
	repo := &mockPickupRepoWorker{}
	w := NewOrganicCancelWorker(repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Start should exit cleanly without panic when context is already done
	w.Start(ctx) // no assertion - just verifying no panic
}

func TestOrganicCancelWorker_Start_TickerFires(t *testing.T) {
	called := make(chan struct{}, 1)
	repo := &mockPickupRepoWorker{
		cancelOrganicPendingFn: func(olderThan time.Duration) (int64, error) {
			called <- struct{}{}
			return 5, nil
		},
	}
	w := NewOrganicCancelWorker(repo)
	w.interval = 10 * time.Millisecond // override for fast test

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Start(ctx)

	select {
	case <-called:
		// ticker fired and CancelOrganicPending was called with affected > 0
	case <-time.After(5 * time.Second):
		t.Fatal("ticker did not fire within timeout")
	}
}

func TestOrganicCancelWorker_Start_TickerFires_Error(t *testing.T) {
	called := make(chan struct{}, 1)
	repo := &mockPickupRepoWorker{
		cancelOrganicPendingFn: func(olderThan time.Duration) (int64, error) {
			called <- struct{}{}
			return 0, errors.New("db error")
		},
	}
	w := NewOrganicCancelWorker(repo)
	w.interval = 10 * time.Millisecond // override for fast test

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Start(ctx)

	select {
	case <-called:
		// ticker fired and CancelOrganicPending returned an error
	case <-time.After(5 * time.Second):
		t.Fatal("ticker did not fire within timeout")
	}
}
