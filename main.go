package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"iletken/config"
	"iletken/logger"
	"iletken/redirector"

	"github.com/valyala/fasthttp"
)

const (
	DefaultConfigPath = "./iletken.yml"
	AppName          = "iletken"
	Version          = "1.0.0"
)

func main() {
	// Parse command line parameters
	configPath := flag.String("config", DefaultConfigPath, "Configuration file path")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		log.Printf("%s v%s", AppName, Version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logger
	logger := logger.NewLogger(cfg.Logging)
	
	logger.Info("Starting İletken",
		slog.String("version", Version),
		slog.String("config_path", *configPath),
		slog.String("listen_address", cfg.Server.GetAddress()),
		slog.Int("redirect_rules", len(cfg.Redirects)),
	)

	// Create redirect handler
	handler := redirector.NewRedirectHandler(cfg.Redirects, logger)

	// Log statistics
	stats := handler.GetStats()
	logger.Info("Redirector ready",
		slog.Any("stats", stats),
	)

	// Configure FastHTTP server
	server := &fasthttp.Server{
		Handler: handler.Handle,
		Name:    AppName + "/" + Version,
	}

	// Set timeouts
	if readTimeout, err := cfg.Server.GetReadTimeout(); err == nil {
		server.ReadTimeout = readTimeout
	}
	if writeTimeout, err := cfg.Server.GetWriteTimeout(); err == nil {
		server.WriteTimeout = writeTimeout
	}
	if idleTimeout, err := cfg.Server.GetIdleTimeout(); err == nil {
		server.IdleTimeout = idleTimeout
	}

	// Create channel for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in background
	go func() {
		logger.Info("Starting HTTP server",
			slog.String("address", cfg.Server.GetAddress()),
		)
		
		if err := server.ListenAndServe(cfg.Server.GetAddress()); err != nil {
			logger.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("İletken ready - waiting for HTTP requests")

	// Wait for shutdown signal
	<-shutdown
	logger.Info("Shutdown signal received, stopping server...")

	// Shutdown server
	if err := server.Shutdown(); err != nil {
		logger.Error("Error during server shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("İletken stopped successfully")
}
