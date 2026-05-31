# saga-orchestrator

A distributed Saga Pattern orchestrator built from scratch in Go — no frameworks, no magic libraries.

Implements the choreography of local transactions with automatic compensation on failure, persistent state in PostgreSQL, and structured JSON logging.

> Built as part of a backend portfolio. Watch the full coding session on [YouTube →](https://www.youtube.com/@FranSammauro)

---

## What is the Saga Pattern?

In distributed systems, you can't wrap multiple services in a single database transaction. The Saga pattern solves this by breaking a workflow into a sequence of local transactions, where each step has a corresponding **compensating transaction** that undoes its work if something fails downstream.

```
Step 1: ReserveStock     → compensation: ReleaseStock
Step 2: ChargePayment    → compensation: RefundPayment
Step 3: CreateShipment   → compensation: CancelShipment
```

If step 3 fails, the orchestrator automatically runs `RefundPayment` and `ReleaseStock` in reverse order. This is not a rollback — it's compensation.

---

## Architecture

```
HTTP Request
     │
     ▼
┌─────────────────────────────┐
│        Orchestrator         │  ← persists state before each step
│                             │    (journal-before-execute pattern)
│  Step 1 → Step 2 → Step 3  │
│     ↓ on failure            │
│  Compensate in reverse      │
└────────────┬────────────────┘
             │
             ▼
        PostgreSQL
   saga_instances + saga_step_log
```

The orchestrator persists every state transition to Postgres before executing a step. This means if the process crashes mid-saga, it can be resumed from where it left off.

---

## Stack

- **Go** — concurrency, performance, no runtime overhead
- **PostgreSQL** — persistent saga state and step audit log
- **Docker** — containerized Postgres with automatic migrations
- **`database/sql` + `lib/pq`** — no ORM, raw SQL

---

## Project Structure

```
saga-orchestrator/
├── cmd/
│   └── server/
│       └── main.go              ← HTTP server + saga registration
├── internal/
│   ├── saga/
│   │   ├── definition.go        ← Step and Definition types
│   │   ├── state.go             ← Instance and Status types
│   │   └── orchestrator.go      ← core: execute and compensate
│   ├── steps/
│   │   ├── reserve_stock.go
│   │   ├── charge_payment.go
│   │   └── create_shipment.go
│   └── store/
│       └── postgres.go          ← persistence layer
├── db/
│   └── migrations/
│       └── 001_create_sagas.sql
├── docker-compose.yml
└── go.mod
```

---

## Getting Started

**Requirements:** Go 1.22+, Docker

```bash
# Clone the repo
git clone https://github.com/FranSammauro/saga-orchestrator
cd saga-orchestrator

# Start Postgres (runs migrations automatically)
docker compose up -d

# Run the server
DATABASE_URL="postgres://saga:saga@localhost:5432/saga_db?sslmode=disable" go run cmd/server/main.go
```

---

## Usage

**Start a saga (happy path):**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"order_id": "ord-001", "items": [{"sku": "LAPTOP", "qty": 1}], "amount": 1500}'

# Response
{"saga_id":"3dc5b520-1b36-4662-83a4-358bdab88651"}
```

**Check saga status:**
```bash
curl http://localhost:8080/orders/3dc5b520-1b36-4662-83a4-358bdab88651/status
```

**Example response (COMPLETED):**
```json
{
  "ID": "3dc5b520-1b36-4662-83a4-358bdab88651",
  "SagaType": "CreateOrder",
  "Status": "COMPLETED",
  "CurrentStep": 2,
  "Payload": {
    "order_id": "ord-001",
    "stock_reservation_id": "res-ord-001",
    "payment_id": "pay-ord-001",
    "shipment_id": "ship-ord-001",
    "tracking": "TRACK-ord-001"
  }
}
```

**Simulate a failure (triggers compensation):**

In `cmd/server/main.go`, change `NewPaymentStep(logger, false)` to `NewPaymentStep(logger, true)` and restart. The orchestrator will execute `ReserveStock`, fail on `ChargePayment`, and automatically compensate by running `ReleaseStock` in reverse.

Server logs during compensation:
```json
{"msg":"executing step","step":"ReservarStock"}
{"msg":"executing step","step":"CobrarPago"}
{"msg":"step failed","step":"CobrarPago","err":"tarjeta rechazada: fondos insuficientes"}
{"msg":"compensating step","step":"ReservarStock"}
```

---

## Key Design Decisions

**Why Lua scripts in Redis for the rate limiter (sister project)?**
Atomic execution prevents race conditions across multiple app instances.

**Why journal-before-execute?**
Persisting the saga state before running a step guarantees the orchestrator can recover from crashes without re-running already-completed steps.

**Why not use a framework like Temporal?**
This is an intentional from-scratch implementation to understand the underlying mechanics. In production, Temporal or a similar framework would be preferred for durability guarantees.

**What happens when compensation fails?**
Currently logged as a critical error. In production, this would be routed to a dead letter queue for manual intervention — the hardest edge case of the Saga pattern.

---

## Saga States

```
STARTED → RUNNING → COMPLETED
                 ↘
              COMPENSATING → FAILED
```

---

## License

MIT
