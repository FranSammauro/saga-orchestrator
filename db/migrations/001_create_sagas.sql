CREATE TYPE saga_status AS ENUM (
  'STARTED', 'RUNNING', 'COMPLETED', 'COMPENSATING', 'FAILED'
);

CREATE TABLE saga_instances (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  saga_type     TEXT NOT NULL,
  status        saga_status NOT NULL DEFAULT 'STARTED',
  current_step  INT NOT NULL DEFAULT 0,
  payload       JSONB NOT NULL,
  result        JSONB,
  error_msg     TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE saga_step_log (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  saga_id       UUID REFERENCES saga_instances(id),
  step_index    INT NOT NULL,
  step_name     TEXT NOT NULL,
  status        TEXT NOT NULL,   -- 'executed', 'compensated', 'failed'
  executed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);