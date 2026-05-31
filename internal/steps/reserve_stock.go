package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/FranSammauro/saga-orchestrator/internal/saga"
)

type StockService struct {
	log *slog.Logger
}

func NewStockStep(log *slog.Logger) saga.Step {
	svc := &StockService{log: log}
	return saga.Step{
		Name:       "ReservarStock",
		Execute:    svc.reserve,
		Compensate: svc.release,
	}
}

func (s *StockService) reserve(ctx context.Context, payload map[string]any) (map[string]any, error) {
	orderID := payload["order_id"].(string)
	s.log.Info("reserving stock", "order", orderID)

	return map[string]any{
		"stock_reservation_id": fmt.Sprintf("res-%s", orderID),
	}, nil
}

func (s *StockService) release(ctx context.Context, payload map[string]any) error {
	reservationID, ok := payload["stock_reservation_id"].(string)
	if !ok {
		return nil
	}
	s.log.Info("releasing stock reservation", "reservation", reservationID)
	return nil
}
