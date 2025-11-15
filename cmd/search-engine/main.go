// Copyright 2013 Web Research Properties, LLC and Matt Wells and Gigablast, Inc.
// Copyright 2025 Open Source Search Engine Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meftunca/open-source-search-engine/pkg/api"
	"github.com/meftunca/open-source-search-engine/pkg/config"
	"github.com/meftunca/open-source-search-engine/pkg/crawler"
	"github.com/meftunca/open-source-search-engine/pkg/database"
	"github.com/meftunca/open-source-search-engine/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init()
	log := logger.GetLogger()
	
	log.Info("Starting Open Source Search Engine...")

	// Load configuration
	cfg := config.Load()
	log.WithFields(map[string]interface{}{
		"db_path": cfg.DatabasePath,
		"port":    cfg.HTTPPort,
		"workers": cfg.CrawlerWorkers,
	}).Info("Configuration loaded")

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()
	log.Info("Database initialized")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start crawler in background
	crawlerInstance := crawler.New(db, cfg)
	go func() {
		if err := crawlerInstance.Start(ctx); err != nil {
			log.WithError(err).Error("Crawler error")
		}
	}()
	log.Info("Crawler started")

	// Setup HTTP server
	apiServer := api.New(db)
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      apiServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.WithField("port", cfg.HTTPPort).Info("HTTP server listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("HTTP server error")
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down gracefully...")

	// Cancel context to stop crawler
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("HTTP server shutdown error")
	}

	log.Info("Shutdown complete")
	fmt.Println("Thank you for using Open Source Search Engine!")
}
