package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/FranSammauro/saga-orchestrator/internal/saga"
)

type ShipmentService struct {
	log *slog.Logger
}

func NewShipmentStep(log *slog.Logger) saga.Step {
	svc := &ShipmentService{log: log}
	return saga.Step{
		Name:       "CrearEnvio",
		Execute:    svc.create,
		Compensate: svc.cancel,
	}
}

func (s *ShipmentService) create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	orderID := payload["order_id"].(string)
	s.log.Info("creating shipment", "order", orderID)

	return map[string]any{
		"shipment_id": fmt.Sprintf("ship-%s", orderID),
		"tracking":    fmt.Sprintf("TRACK-%s", orderID),
	}, nil
}

func (s *ShipmentService) cancel(ctx context.Context, payload map[string]any) error {
	shipmentID, ok := payload["shipment_id"].(string)
	if !ok {
		return nil
	}
	s.log.Info("cancelling shipment", "shipment", shipmentID)
	return nil
}
