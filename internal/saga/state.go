package saga

import "time"

type Status string

const (
    StatusStarted      Status = "STARTED"
    StatusRunning      Status = "RUNNING"
    StatusCompleted    Status = "COMPLETED"
    StatusCompensating Status = "COMPENSATING"
    StatusFailed       Status = "FAILED"
)

type Instance struct {
    ID          string
    SagaType    string
    Status      Status
    CurrentStep int
    Payload     map[string]any
    Result      map[string]any
    ErrorMsg    string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}