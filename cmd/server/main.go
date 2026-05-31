package main

import (
    "context"
    "encoding/json"
    "log/slog"
    "net/http"
    "os"

    "github.com/FranSammauro/saga-orchestrator/internal/saga"
    "github.com/FranSammauro/saga-orchestrator/internal/steps"
    "github.com/FranSammauro/saga-orchestrator/internal/store"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    pg, err := store.NewPostgres(os.Getenv("DATABASE_URL"))
    if err != nil {
        logger.Error("connecting to postgres", "err", err)
        os.Exit(1)
    }

    orchestrator := saga.NewOrchestrator(pg, logger)

    orchestrator.Register(&saga.Definition{
        Name: "CreateOrder",
        Steps: []saga.Step{
            steps.NewStockStep(logger),
            steps.NewPaymentStep(logger, true),
            steps.NewShipmentStep(logger),
        },
    })

    http.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
        var payload map[string]any
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }

        sagaID, err := orchestrator.Start(context.Background(), "CreateOrder", payload)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusAccepted) 
        json.NewEncoder(w).Encode(map[string]string{"saga_id": sagaID})
    })

    http.HandleFunc("GET /orders/{id}/status", func(w http.ResponseWriter, r *http.Request) {
        inst, err := orchestrator.GetStatus(r.Context(), r.PathValue("id"))
        if err != nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(inst)
    })

    logger.Info("server listening", "port", 8080)
    http.ListenAndServe(":8080", nil)
}