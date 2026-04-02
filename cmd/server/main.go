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
	"github.com/truggeri/go-garage/internal/auth"
	"github.com/truggeri/go-garage/internal/config"
	"github.com/truggeri/go-garage/internal/database"
	"github.com/truggeri/go-garage/internal/handlers"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
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

	// Initialize repositories
	db := garageDB.RawSQLConnection()
	userRepo := repositories.NewSQLiteUserRepository(db)
	vehicleRepo := repositories.NewSQLiteVehicleRepository(db)
	maintenanceRepo := repositories.NewSQLiteMaintenanceRepository(db)
	fuelRepo := repositories.NewSQLiteFuelRepository(db)
	metricsRepo := repositories.NewSQLiteMetricsRepository(db)

	// Initialize services
	userSvc := services.NewUserService(userRepo)
	vehicleSvc := services.NewVehicleService(vehicleRepo)
	maintenanceSvc := services.NewMaintenanceService(maintenanceRepo, vehicleRepo, metricsRepo)
	fuelSvc := services.NewFuelService(fuelRepo, vehicleRepo, metricsRepo)
	_ = fuelSvc // TODO: wire into handlers when fuel API/page endpoints are added

	// Initialize JWT token manager
	tokenMgr, tokenErr := auth.BuildTokenManager(cfg.JWT.Secret, auth.StandardTokenDurations())
	if tokenErr != nil {
		vehicleLog.RecordError("Failed to initialize token manager", "error", tokenErr.Error())
		os.Exit(1)
	}

	// Initialize authentication service
	authSvc := services.BuildAuthService(userRepo, tokenMgr)

	// Initialize handlers
	authHandler := handlers.BuildAuthHandler(authSvc)
	vehicleHandler := handlers.MakeVehicleAPIHandler(vehicleSvc)
	maintenanceHandler := handlers.MakeMaintenanceAPIHandler(maintenanceSvc, vehicleSvc)
	userHandler := handlers.MakeUserAPIHandler(userSvc)

	// Initialize template engine
	tmplEngine := templateengine.NewEngine("./web/templates", cfg.IsDevelopment())
	if !cfg.IsDevelopment() {
		if err := tmplEngine.LoadTemplates(); err != nil {
			vehicleLog.RecordError("Failed to load templates", "error", err.Error())
			os.Exit(1)
		}
	}
	pageHandler := handlers.NewPageHandler(tmplEngine, authSvc, vehicleSvc, maintenanceSvc, userSvc, metricsRepo)

	// Setup router and routes
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(pageHandler.NotFound)

	// Static file serving
	staticDir := "./web/static/"
	staticFileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	router.PathPrefix("/static/").Handler(staticFileServer)

	// Health check endpoint (no auth required)
	router.HandleFunc("/health", createHealthCheckHandler(garageDB)).Methods("GET")

	// Web page routes (public, with CSRF protection for forms)
	publicPages := router.NewRoute().Subrouter()
	publicPages.Use(middleware.CSRFProtection(cfg.CSRF.Secret))
	publicPages.HandleFunc("/", pageHandler.Home).Methods("GET")
	publicPages.HandleFunc("/register", pageHandler.RegisterForm).Methods("GET")
	publicPages.HandleFunc("/register", pageHandler.RegisterSubmit).Methods("POST")
	publicPages.HandleFunc("/login", pageHandler.LoginForm).Methods("GET")
	publicPages.HandleFunc("/login", pageHandler.LoginSubmit).Methods("POST")

	// Web page routes (require cookie authentication + CSRF protection)
	// CookieAuthGuard must run before CSRFProtection so that session info
	// (AccountInfo) is available in context for HMAC-based token generation.
	protectedPages := router.NewRoute().Subrouter()
	protectedPages.Use(middleware.CookieAuthGuard(tokenMgr))
	protectedPages.Use(middleware.CSRFProtection(cfg.CSRF.Secret))
	protectedPages.HandleFunc("/dashboard", pageHandler.Dashboard).Methods("GET")
	protectedPages.HandleFunc("/vehicles", pageHandler.VehicleList).Methods("GET")
	protectedPages.HandleFunc("/vehicles/new", pageHandler.VehicleNew).Methods("GET")
	protectedPages.HandleFunc("/vehicles/new", pageHandler.VehicleCreate).Methods("POST")
	protectedPages.HandleFunc("/maintenance", pageHandler.MaintenanceList).Methods("GET")
	protectedPages.HandleFunc("/maintenance/new", pageHandler.MaintenanceNew).Methods("GET")
	protectedPages.HandleFunc("/maintenance/new", pageHandler.MaintenanceCreate).Methods("POST")
	protectedPages.HandleFunc("/maintenance/{id}", pageHandler.MaintenanceDetail).Methods("GET")
	protectedPages.HandleFunc("/maintenance/{id}/edit", pageHandler.MaintenanceEdit).Methods("GET")
	protectedPages.HandleFunc("/maintenance/{id}/edit", pageHandler.MaintenanceUpdate).Methods("POST")
	protectedPages.HandleFunc("/profile", pageHandler.ViewProfile).Methods("GET")
	protectedPages.HandleFunc("/profile/edit", pageHandler.ProfileEdit).Methods("GET")
	protectedPages.HandleFunc("/profile/edit", pageHandler.ProfileUpdate).Methods("POST")
	protectedPages.HandleFunc("/profile/password", pageHandler.ChangePassword).Methods("GET")
	protectedPages.HandleFunc("/profile/password", pageHandler.ChangePasswordSubmit).Methods("POST")

	// Vehicle detail page routes (require cookie auth + vehicle ownership)
	vehiclePages := protectedPages.PathPrefix("/vehicles/{id}").Subrouter()
	vehiclePages.Use(middleware.PageResourceOwnershipGuard(newVehicleLookup(vehicleSvc), pageHandler.RenderError))
	vehiclePages.HandleFunc("", pageHandler.VehicleDetail).Methods("GET")
	vehiclePages.HandleFunc("/edit", pageHandler.VehicleEdit).Methods("GET")
	vehiclePages.HandleFunc("/edit", pageHandler.VehicleUpdate).Methods("POST")

	// API v1 routes
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Auth routes (no auth required)
	apiV1.HandleFunc("/auth/register", authHandler.HandleRegister).Methods("POST")
	apiV1.HandleFunc("/auth/login", authHandler.HandleLogin).Methods("POST")
	apiV1.HandleFunc("/auth/refresh", authHandler.HandleRefresh).Methods("POST")
	apiV1.HandleFunc("/auth/logout", authHandler.HandleLogout).Methods("POST")

	// Protected routes (require authentication)
	protected := apiV1.NewRoute().Subrouter()
	protected.Use(middleware.AuthenticationGuard(tokenMgr))

	// Vehicle routes
	protected.HandleFunc("/vehicles", vehicleHandler.ListAll).Methods("GET")
	protected.HandleFunc("/vehicles", vehicleHandler.CreateOne).Methods("POST")
	protected.HandleFunc("/vehicles/{id}", vehicleHandler.GetOne).Methods("GET")
	protected.HandleFunc("/vehicles/{id}", vehicleHandler.ReplaceOne).Methods("PUT")
	protected.HandleFunc("/vehicles/{id}", vehicleHandler.RemoveOne).Methods("DELETE")
	protected.HandleFunc("/vehicles/{id}/stats", vehicleHandler.GetStats).Methods("GET")

	// Maintenance routes
	protected.HandleFunc("/vehicles/{vehicleId}/maintenance", maintenanceHandler.ListAll).Methods("GET")
	protected.HandleFunc("/vehicles/{vehicleId}/maintenance", maintenanceHandler.CreateOne).Methods("POST")
	protected.HandleFunc("/maintenance/{id}", maintenanceHandler.GetOne).Methods("GET")
	protected.HandleFunc("/maintenance/{id}", maintenanceHandler.ReplaceOne).Methods("PUT")
	protected.HandleFunc("/maintenance/{id}", maintenanceHandler.RemoveOne).Methods("DELETE")

	// User routes
	protected.HandleFunc("/users/me", userHandler.GetMe).Methods("GET")
	protected.HandleFunc("/users/me", userHandler.UpdateMe).Methods("PUT")
	protected.HandleFunc("/users/me", userHandler.DeleteMe).Methods("DELETE")
	protected.HandleFunc("/users/me/password", userHandler.ChangePassword).Methods("PUT")

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
		fmt.Fprintf(w, `{"status":"%s","database":"%s","timestamp":"%s"}`,
			overallStatus, dbStatus, time.Now().Format(time.RFC3339))
	}
}
