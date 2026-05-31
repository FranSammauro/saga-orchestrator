package steps

import (
    "context"
    "errors"
    "fmt"
    "log/slog"

    "github.com/FranSammauro/saga-orchestrator/internal/saga"
)

type PaymentService struct {
    log        *slog.Logger
    shouldFail bool
}

func NewPaymentStep(log *slog.Logger, simulateFail bool) saga.Step {
    svc := &PaymentService{log: log, shouldFail: simulateFail}
    return saga.Step{
        Name:       "CobrarPago",
        Execute:    svc.charge,
        Compensate: svc.refund,
    }
}

func (p *PaymentService) charge(ctx context.Context, payload map[string]any) (map[string]any, error) {
    if p.shouldFail {
        return nil, errors.New("tarjeta rechazada: fondos insuficientes")
    }

    orderID := payload["order_id"].(string)
    p.log.Info("charging payment", "order", orderID)

    return map[string]any{
        "payment_id": fmt.Sprintf("pay-%s", orderID),
    }, nil
}

func (p *PaymentService) refund(ctx context.Context, payload map[string]any) error {
    paymentID, ok := payload["payment_id"].(string)
    if !ok {
        return nil
    }
    p.log.Info("refunding payment", "payment", paymentID)

    return nil
}