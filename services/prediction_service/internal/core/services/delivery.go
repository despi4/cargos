package services

import (
	"context"
	"errors"
	"fmt"
)

// The saved prediction is the outbox. Acknowledgement follows receiver commit;
// event-ID deduplication makes a crash between those steps safe to retry.
func (s *Predictions) DeliverPending(ctx context.Context) error {
	pending, err := s.repository.Pending(ctx)
	if err != nil {
		return err
	}
	var failures []error
	for _, v := range pending {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		message := fmt.Sprintf("%d products need replenishment. Prediction: %s", v.Count, v.PredictionID)
		if err = s.notifier.Publish(ctx, v.OwnerID, v.PredictionID, "Stock replenishment required", message); err != nil {
			failures = append(failures, err)
			continue
		}
		if err = s.repository.MarkSent(ctx, v.OwnerID, v.PredictionID); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
