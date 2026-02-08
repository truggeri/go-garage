package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/config"
	"github.com/truggeri/go-garage/internal/database"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/pkg/applog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	vehicleLog := applog.BuildVehicleAppLog(cfg.Logging.Level, cfg.Logging.Format, nil)
	vehicleLog.RecordAppStartup(cfg.Env, cfg.Server.Host, cfg.Server.Port)

	// Initialize database connection
	vehicleLog.RecordInfo("Establishing database connection", "path", cfg.Database.Path)
	garageDB, err := database.InitializeGarage(cfg.Database.Path, database.StandardWorkerPoolSettings())
	if err != nil {
		vehicleLog.RecordError("Failed to establish database connection", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := garageDB.Terminate(); err != nil {
			vehicleLog.RecordError("Failed to close database connection", "error", err.Error())
		}
	}()

	// Verify database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := garageDB.DiagnoseHealth(ctx); err != nil {
		cancel()
		vehicleLog.RecordError("Database connectivity check failed", "error", err.Error())
		os.Exit(1)
	}
	cancel()
	vehicleLog.RecordInfo("Database connectivity verified")

	// Run database migrations
	vehicleLog.RecordInfo("Running database migrations")
	migrationsPath := "./migrations"
	if err := database.BootstrapSchema(context.Background(), garageDB, migrationsPath); err != nil {
		vehicleLog.RecordError("Failed to run database migrations", "error", err.Error())
		os.Exit(1)
	}
	vehicleLog.RecordInfo("Database migrations completed successfully")

	router := mux.NewRouter()
	router.HandleFunc("/health", createHealthCheckHandler(garageDB)).Methods("GET")

	handler := middleware.RequestLogger(vehicleLog)(router)
	handler = middleware.RecoverFromPanic(vehicleLog)(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		vehicleLog.RecordAppShutdown(fmt.Sprintf("Signal %v received", sig))

		if err := server.Close(); err != nil {
			vehicleLog.RecordError("Server forced to shutdown", "error", err.Error())
			os.Exit(1)
		}

		close(done)
	}()

	vehicleLog.RecordInfo("Server listening", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		vehicleLog.RecordError("Server failed to start", "error", err.Error())
		os.Exit(1)
	}

	<-done
	vehicleLog.RecordInfo("Server stopped successfully")
}

// createHealthCheckHandler returns a handler that includes database health check
func createHealthCheckHandler(garageDB *database.SQLiteGarage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "healthy"
		overallStatus := "healthy"
		statusCode := http.StatusOK
		
		if err := garageDB.DiagnoseHealth(ctx); err != nil {
			dbStatus = "unhealthy"
			overallStatus = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		//nolint:errcheck
		fmt.Fprintf(w, `{"status":"%s","database":"%s","timestamp":"%s"}`,
			overallStatus, dbStatus, time.Now().Format(time.RFC3339))
	}
}
