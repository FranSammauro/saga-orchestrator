package saga

import (
    "context"
    "fmt"
    "log/slog"
)

type Store interface {
    CreateInstance(ctx context.Context, inst *Instance) error
    UpdateInstance(ctx context.Context, inst *Instance) error
    GetInstance(ctx context.Context, id string) (*Instance, error)
    LogStep(ctx context.Context, sagaID string, stepIdx int, name, status string) error
}

type Orchestrator struct {
    store Store
    defs  map[string]*Definition
    log   *slog.Logger
}

func NewOrchestrator(store Store, logger *slog.Logger) *Orchestrator {
    return &Orchestrator{
        store: store,
        defs:  make(map[string]*Definition),
        log:   logger,
    }
}

func (o *Orchestrator) Register(def *Definition) {
    o.defs[def.Name] = def
}

func (o *Orchestrator) Start(ctx context.Context, sagaType string, payload map[string]any) (string, error) {
    def, ok := o.defs[sagaType]
    if !ok {
        return "", fmt.Errorf("saga type %q not registered", sagaType)
    }

    inst := &Instance{
        SagaType:    sagaType,
        Status:      StatusStarted,
        CurrentStep: 0,
        Payload:     payload,
    }

    if err := o.store.CreateInstance(ctx, inst); err != nil {
        return "", fmt.Errorf("creating instance: %w", err)
    }

    go o.run(context.Background(), inst, def)
    return inst.ID, nil
}

func (o *Orchestrator) run(ctx context.Context, inst *Instance, def *Definition) {
    inst.Status = StatusRunning
    _ = o.store.UpdateInstance(ctx, inst)

    executedSteps := []int{}

    for i, step := range def.Steps {
        inst.CurrentStep = i
        o.log.Info("executing step", "saga", inst.ID, "step", step.Name)

        result, err := step.Execute(ctx, inst.Payload)
        if err != nil {
            o.log.Error("step failed", "step", step.Name, "err", err)
            _ = o.store.LogStep(ctx, inst.ID, i, step.Name, "failed")

            o.compensate(ctx, inst, def, executedSteps)
            inst.Status = StatusFailed
            inst.ErrorMsg = fmt.Sprintf("step %q failed: %v", step.Name, err)
            _ = o.store.UpdateInstance(ctx, inst)
            return
        }

    
        for k, v := range result {
            inst.Payload[k] = v
        }

        _ = o.store.LogStep(ctx, inst.ID, i, step.Name, "executed")
        executedSteps = append(executedSteps, i)
    }

    inst.Status = StatusCompleted
    inst.Result = inst.Payload
    _ = o.store.UpdateInstance(ctx, inst)
    o.log.Info("saga completed", "saga", inst.ID)
}


func (o *Orchestrator) compensate(ctx context.Context, inst *Instance, def *Definition, executedSteps []int) {
    inst.Status = StatusCompensating
    _ = o.store.UpdateInstance(ctx, inst)

    for i := len(executedSteps) - 1; i >= 0; i-- {
        stepIdx := executedSteps[i]
        step := def.Steps[stepIdx]

        o.log.Info("compensating step", "saga", inst.ID, "step", step.Name)

        if err := step.Compensate(ctx, inst.Payload); err != nil {
            // En producción, esto iría a una dead letter queue para intervención manual.
            // La compensación fallida es el caso más difícil del patrón Saga.
            o.log.Error("compensation failed — needs manual intervention",
                "step", step.Name, "err", err)
        }

        _ = o.store.LogStep(ctx, inst.ID, stepIdx, step.Name, "compensated")
    }
}
func (o *Orchestrator) GetStatus(ctx context.Context, id string) (*Instance, error) {
	return o.store.GetInstance(ctx, id)
}
