package saga

import "context"


type Step struct {
    Name       string
    Execute    func(ctx context.Context, payload map[string]any) (map[string]any, error)
    Compensate func(ctx context.Context, payload map[string]any) error
}


type Definition struct {
    Name  string
    Steps []Step
}