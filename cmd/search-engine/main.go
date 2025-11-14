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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meftunca/open-source-search-engine/pkg/api"
	"github.com/meftunca/open-source-search-engine/pkg/config"
	"github.com/meftunca/open-source-search-engine/pkg/crawler"
	"github.com/meftunca/open-source-search-engine/pkg/database"
)

func main() {
	log.Println("Starting Open Source Search Engine...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration: DB=%s, Port=%s, Workers=%d\n", 
		cfg.DatabasePath, cfg.HTTPPort, cfg.CrawlerWorkers)

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start crawler in background
	crawlerInstance := crawler.New(db, cfg)
	go func() {
		if err := crawlerInstance.Start(ctx); err != nil {
			log.Printf("Crawler error: %v", err)
		}
	}()
	log.Println("Crawler started")

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
		log.Printf("HTTP server listening on :%s\n", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down gracefully...")

	// Cancel context to stop crawler
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
	fmt.Println("Thank you for using Open Source Search Engine!")
}
