package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"tekton-hub-proxy/internal/config"
	"tekton-hub-proxy/internal/handlers"
	"tekton-hub-proxy/internal/client"
	"tekton-hub-proxy/internal/translator"
)

func main() {
	// Parse command line flags
	var (
		debug      = flag.Bool("debug", false, "Enable debug logging")
		configPath = flag.String("config", "", "Path to config file")
		port       = flag.Int("port", 0, "Server port (overrides config)")
		bindAddr   = flag.String("bind", "", "Bind address (overrides config)")
		help       = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Tekton Hub to Artifact Hub Translation Proxy")
		fmt.Println()
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadWithPath(*configPath)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command line flags
	if *debug {
		cfg.Logging.Level = "debug"
	}
	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *bindAddr != "" {
		cfg.Server.Host = *bindAddr
	}

	// Setup logging
	setupLogging(cfg.Logging)

	logrus.WithField("config", cfg).Info("Starting Tekton Hub Proxy")

	// Create Artifact Hub client
	artifactHubClient := client.NewArtifactHubClient(cfg.ArtifactHub)

	// Create translator
	catalogTranslator := translator.NewCatalogTranslator(cfg.CatalogMappings)
	responseTranslator := translator.NewResponseTranslator()
	versionTranslator := translator.NewVersionTranslator()

	// Create handlers
	handlers := handlers.NewHandlers(
		artifactHubClient,
		catalogTranslator,
		responseTranslator,
		versionTranslator,
	)

	// Setup routes
	router := setupRoutes(handlers)

	// Add middleware
	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.CORSMiddleware)
	router.Use(handlers.RecoveryMiddleware)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logrus.WithField("address", addr).Info("Server starting")

	if err := http.ListenAndServe(addr, router); err != nil {
		logrus.Fatalf("Server failed to start: %v", err)
	}
}

func setupLogging(cfg config.LoggingConfig) {
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		logrus.Warnf("Invalid log level %q, using info", cfg.Level)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if cfg.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.SetOutput(os.Stdout)
}

func setupRoutes(h *handlers.Handlers) *mux.Router {
	router := mux.NewRouter()

	// Catalog endpoints
	router.HandleFunc("/v1/catalogs", h.ListCatalogs).Methods("GET")

	// Resource endpoints (order matters - more specific routes first)
	router.HandleFunc("/v1/resource/{catalog}/{kind}/{name}/raw", h.GetLatestResourceYAML).Methods("GET")
	router.HandleFunc("/v1/resource/{catalog}/{kind}/{name}/{version}/yaml", h.GetResourceYAML).Methods("GET")
	router.HandleFunc("/v1/resource/{catalog}/{kind}/{name}/{version}/readme", h.GetResourceReadme).Methods("GET")
	router.HandleFunc("/v1/resource/{catalog}/{kind}/{name}/{version}/raw", h.GetResourceYAMLRaw).Methods("GET")
	router.HandleFunc("/v1/resource/{catalog}/{kind}/{name}/{version}", h.GetResourceVersion).Methods("GET")
	router.HandleFunc("/v1/resource/{catalog}/{kind}/{name}", h.GetResource).Methods("GET")
	router.HandleFunc("/v1/resource/{id:[0-9]+}", h.GetResourceByID).Methods("GET")
	router.HandleFunc("/v1/resource/{id:[0-9]+}/versions", h.GetResourceVersionsByID).Methods("GET")
	router.HandleFunc("/v1/resource/version/{versionID:[0-9]+}", h.GetResourceByVersionID).Methods("GET")

	// List and query endpoints
	router.HandleFunc("/v1/resources", h.ListResources).Methods("GET")
	router.HandleFunc("/v1/query", h.QueryResources).Methods("GET")

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	return router
}