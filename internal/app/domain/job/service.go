package job

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

const (
	workerCount  = 10
	workerBuffer = 10
)

type Repository interface {
	Save(ctx context.Context, j *Job) error
	GetByChannelID(ctx context.Context, chID uuid.UUID) ([]*Job, error)
}

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, r Repository, jobs <-chan *Job, errs chan<- error) {
	for j := range jobs {
		if err := r.Save(ctx, j); err != nil {
			errs <- fmt.Errorf("failed to save job %s: %w", j.ID(), err)
		}
	}
	wg.Done()
}

func (s *Service) Sync(ctx context.Context, chID uuid.UUID, incoming []*Job) error {
	// get existing jobs
	existing, err := s.r.GetByChannelID(ctx, chID)
	if err != nil {
		return fmt.Errorf("failed to get existing jobs: %w", err)
	}

	// create job workers
	var wgWorkers sync.WaitGroup
	jobs := make(chan *Job, workerBuffer)
	errs := make(chan error, workerBuffer)
	for w := 1; w <= workerCount; w++ {
		wgWorkers.Add(1)
		go worker(ctx, &wgWorkers, s.r, jobs, errs)
	}

	// create error worker
	var syncErrs error
	var wgError sync.WaitGroup
	wgError.Add(1)
	go func(errs <-chan error) {
		for err := range errs {
			syncErrs = errors.Join(syncErrs, err)
		}
		wgError.Done()
	}(errs)

	// save if incoming does not exist or is different
	for _, in := range incoming {
		for _, ex := range existing {
			if ex.ID() == in.ID() {
				if in.IsEqual(ex) {
					goto next
				}
			}
		}

		in.MarkAsChanged()
		jobs <- in

	next:
	}

	// save if existing does not exist in incoming
	for _, ex := range existing {
		for _, in := range incoming {
			if ex.ID() == in.ID() {
				goto skip
			}
		}

		ex.MarkAsMissing()
		jobs <- ex

	skip:
	}
	close(jobs)
	wgWorkers.Wait()
	close(errs)
	wgError.Wait()

	return syncErrs
}
