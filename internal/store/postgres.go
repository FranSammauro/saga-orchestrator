package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FranSammauro/saga-orchestrator/internal/saga"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(connStr string) (*Postgres, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	return &Postgres{db: db}, nil
}

func (p *Postgres) CreateInstance(ctx context.Context, inst *saga.Instance) error {
	payload, _ := json.Marshal(inst.Payload)
	err := p.db.QueryRowContext(ctx, `
		INSERT INTO saga_instances (saga_type, status, current_step, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		inst.SagaType, inst.Status, inst.CurrentStep, payload,
	).Scan(&inst.ID, &inst.CreatedAt, &inst.UpdatedAt)
	return err
}

func (p *Postgres) UpdateInstance(ctx context.Context, inst *saga.Instance) error {
	payload, _ := json.Marshal(inst.Payload)
	result, _ := json.Marshal(inst.Result)
	inst.UpdatedAt = time.Now()
	_, err := p.db.ExecContext(ctx, `
		UPDATE saga_instances
		SET status=$1, current_step=$2, payload=$3, result=$4, error_msg=$5, updated_at=$6
		WHERE id=$7`,
		inst.Status, inst.CurrentStep, payload, result, inst.ErrorMsg, inst.UpdatedAt, inst.ID,
	)
	return err
}

func (p *Postgres) GetInstance(ctx context.Context, id string) (*saga.Instance, error) {
	inst := &saga.Instance{}
	var payload, result []byte
	var errorMsg sql.NullString
	err := p.db.QueryRowContext(ctx,
		`SELECT id, saga_type, status, current_step, payload, result, error_msg, created_at, updated_at
		 FROM saga_instances WHERE id=$1`, id,
	).Scan(&inst.ID, &inst.SagaType, &inst.Status, &inst.CurrentStep,
		&payload, &result, &errorMsg, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(payload, &inst.Payload)
	json.Unmarshal(result, &inst.Result)
	inst.ErrorMsg = errorMsg.String
	return inst, nil
}

func (p *Postgres) LogStep(ctx context.Context, sagaID string, stepIdx int, name, status string) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO saga_step_log (saga_id, step_index, step_name, status)
		VALUES ($1, $2, $3, $4)`,
		sagaID, stepIdx, name, status,
	)
	return err
}
