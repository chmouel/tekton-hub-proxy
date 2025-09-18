package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

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
		debug             = flag.Bool("debug", false, "Enable debug logging")
		configPath        = flag.String("config", "", "Path to config file")
		port              = flag.Int("port", 0, "Server port (overrides config)")
		bindAddr          = flag.String("bind", "", "Bind address (overrides config)")
		disableLandingPage = flag.Bool("disable-landing-page", false, "Disable the landing page at root path (/)")
		disableCache      = flag.Bool("disable-cache", false, "Disable API response caching")
		cacheTTL          = flag.String("cache-ttl", "", "Cache TTL duration (e.g., 5m, 10m) (overrides config)")
		cacheMaxSize      = flag.Int("cache-max-size", 0, "Maximum number of cache entries (overrides config)")
		help              = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Tekton Hub to Artifact Hub Translation Proxy")
		fmt.Println()
		fmt.Println("A seamless translation proxy that bridges Tekton Hub API calls to Artifact Hub,")
		fmt.Println("enabling compatibility between systems while leveraging Artifact Hub's powerful catalog.")
		fmt.Println()
		fmt.Println("Usage:")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Configuration:")
		fmt.Println("  Cache settings can be configured via config file:")
		fmt.Println("  artifacthub:")
		fmt.Println("    cache:")
		fmt.Println("      enabled: true")
		fmt.Println("      ttl: 5m")
		fmt.Println("      max_size: 2000")
		fmt.Println()
		fmt.Println("  Or via environment variables:")
		fmt.Println("  THP_ARTIFACTHUB_CACHE_ENABLED=true")
		fmt.Println("  THP_ARTIFACTHUB_CACHE_TTL=5m")
		fmt.Println("  THP_ARTIFACTHUB_CACHE_MAX_SIZE=2000")
		fmt.Println("  THP_LANDING_PAGE_ENABLED=false")
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
	if *disableLandingPage {
		cfg.LandingPage.Enabled = false
	}
	if *disableCache {
		cfg.ArtifactHub.Cache.Enabled = false
	}
	if *cacheTTL != "" {
		if ttl, err := time.ParseDuration(*cacheTTL); err == nil {
			cfg.ArtifactHub.Cache.TTL = ttl
		} else {
			logrus.Warnf("Invalid cache TTL duration %q, using config default", *cacheTTL)
		}
	}
	if *cacheMaxSize > 0 {
		cfg.ArtifactHub.Cache.MaxSize = *cacheMaxSize
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
		cfg,
	)

	// Setup routes
	router := setupRoutes(handlers, cfg)

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

func setupRoutes(h *handlers.Handlers, cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// Landing page (configurable)
	if cfg.LandingPage.Enabled {
		router.HandleFunc("/", h.LandingPage).Methods("GET")
	}

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