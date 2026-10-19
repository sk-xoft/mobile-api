package payment

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		s.runBatch(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *Service) runBatch(ctx context.Context) {
	ids, err := s.store.Due(ctx, 50)
	if err != nil {
		if ctx.Err() == nil {
			zap.L().Error("payment worker scan failed", zap.Error(err))
		}
		return
	}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				if ctx.Err() != nil {
					return
				}
				a, err := s.store.Get(ctx, id)
				if err == nil {
					_, err = s.reconcile(ctx, a, false)
				}
				if err != nil && ctx.Err() == nil {
					zap.L().Error("payment reconciliation failed", zap.String("transaction_id", id), zap.Error(err))
				}
			}
		}()
	}
	for _, id := range ids {
		select {
		case jobs <- id:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		}
	}
	close(jobs)
	wg.Wait()
}
