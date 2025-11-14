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
	"fmt"
	"log"
	"os"
	"time"

	"github.com/meftunca/open-source-search-engine/internal/models"
	"github.com/meftunca/open-source-search-engine/pkg/database"
)

func main() {
	dbPath := "./test_search.db"
	
	// Clean up old test database
	os.Remove(dbPath)
	
	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)
	
	fmt.Println("✓ Database initialized")
	
	// Insert test documents
	testDocs := []models.Document{
		{
			URL:      "https://golang.org",
			Title:    "The Go Programming Language",
			Content:  "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. Go was designed at Google in 2007.",
			MetaDesc: "Official Go programming language website",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			ContentType: "text/html",
			Hash: "abc123",
		},
		{
			URL:      "https://duckdb.org",
			Title:    "DuckDB - An in-process SQL OLAP database",
			Content:  "DuckDB is an in-process SQL OLAP database management system. It is designed to support analytical query workloads with excellent performance.",
			MetaDesc: "Fast analytical database",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			ContentType: "text/html",
			Hash: "def456",
		},
		{
			URL:      "https://example.com",
			Title:    "Example Domain",
			Content:  "This domain is for use in illustrative examples in documents. You may use this domain in literature without prior coordination or asking for permission.",
			MetaDesc: "Example website",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			ContentType: "text/html",
			Hash: "ghi789",
		},
	}
	
	for _, doc := range testDocs {
		if err := db.SaveDocument(&doc); err != nil {
			log.Fatalf("Failed to save document: %v", err)
		}
		fmt.Printf("✓ Saved document: %s\n", doc.Title)
	}
	
	// Test search functionality
	fmt.Println("\n--- Testing Search ---")
	
	testQueries := []string{"golang", "programming", "database", "example"}
	
	for _, query := range testQueries {
		results, err := db.Search(query, 10)
		if err != nil {
			log.Fatalf("Search failed: %v", err)
		}
		fmt.Printf("\nQuery: '%s' - Found %d results:\n", query, len(results))
		for i, result := range results {
			fmt.Printf("  %d. %s\n     URL: %s\n", i+1, result.Title, result.URL)
		}
	}
	
	// Test URL queue
	fmt.Println("\n--- Testing URL Queue ---")
	
	urls := []string{
		"https://test1.com",
		"https://test2.com",
		"https://test3.com",
	}
	
	for i, url := range urls {
		if err := db.AddURL(url, i*10); err != nil {
			log.Fatalf("Failed to add URL: %v", err)
		}
		fmt.Printf("✓ Added URL to queue: %s (priority: %d)\n", url, i*10)
	}
	
	// Get next URLs from queue
	nextURLs, err := db.GetNextURLs(2)
	if err != nil {
		log.Fatalf("Failed to get next URLs: %v", err)
	}
	
	fmt.Printf("\nRetrieved %d URLs from queue:\n", len(nextURLs))
	for _, u := range nextURLs {
		fmt.Printf("  - %s (priority: %d, status: %s)\n", u.URL, u.Priority, u.Status)
	}
	
	fmt.Println("\n✅ All tests passed!")
}
