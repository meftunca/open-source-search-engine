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

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/meftunca/open-source-search-engine/internal/models"
	"github.com/meftunca/open-source-search-engine/pkg/database"
	"github.com/meftunca/open-source-search-engine/pkg/logger"
)

func setupTestDB(t *testing.T) *database.DB {
	// Initialize logger for tests
	logger.Init()
	
	// Create temp database file
	tmpFile := "/tmp/test_search_" + time.Now().Format("20060102150405") + ".db"
	db, err := database.New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	// Clean up on test completion
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})
	
	return db
}

func TestHealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Check that metrics exist
	if _, ok := response["total_requests"]; !ok {
		t.Error("Expected 'total_requests' in metrics response")
	}
}

func TestCrawlEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	reqBody := map[string]interface{}{
		"url":      "https://example.com",
		"priority": 10,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/api/crawl", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestSearchEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	// Add a test document
	doc := &models.Document{
		URL:      "https://example.com/test",
		Title:    "Test Document",
		Content:  "This is a test document about golang programming",
		MetaDesc: "A test document",
		CrawledAt: time.Now(),
		IndexedAt: time.Now(),
		StatusCode: 200,
		Hash:      "testhash123",
	}
	if err := db.SaveDocument(doc); err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}
	
	// Search for the document
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=golang&limit=10", nil)
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["query"] != "golang" {
		t.Errorf("Expected query 'golang', got '%v'", response["query"])
	}
	
	if count, ok := response["count"].(float64); !ok || count < 1 {
		t.Errorf("Expected at least 1 result, got %v", count)
	}
}

func TestImageSearchEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	// Add a test document with images
	doc := &models.Document{
		URL:       "https://example.com/images",
		Title:     "Image Gallery",
		Content:   "A page with images",
		ImageURLs: `["https://example.com/img1.jpg", "https://example.com/img2.jpg"]`,
		CrawledAt: time.Now(),
		IndexedAt: time.Now(),
		StatusCode: 200,
		Hash:      "imagehash123",
	}
	if err := db.SaveDocument(doc); err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}
	
	// Search for images
	req := httptest.NewRequest(http.MethodGet, "/api/search/images?q=images&limit=10", nil)
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["type"] != "images" {
		t.Errorf("Expected type 'images', got '%v'", response["type"])
	}
}

func TestVideoSearchEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	// Add a test document with videos
	doc := &models.Document{
		URL:       "https://example.com/videos",
		Title:     "Video Gallery",
		Content:   "A page with videos",
		VideoURLs: `["https://youtube.com/watch?v=123", "https://vimeo.com/456"]`,
		CrawledAt: time.Now(),
		IndexedAt: time.Now(),
		StatusCode: 200,
		Hash:      "videohash123",
	}
	if err := db.SaveDocument(doc); err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}
	
	// Search for videos
	req := httptest.NewRequest(http.MethodGet, "/api/search/videos?q=videos&limit=10", nil)
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["type"] != "videos" {
		t.Errorf("Expected type 'videos', got '%v'", response["type"])
	}
}

func TestSemanticSearchEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := New(db)
	
	// Add a test document
	doc := &models.Document{
		URL:      "https://example.com/ai",
		Title:    "AI and Machine Learning",
		Content:  "This document discusses artificial intelligence",
		MetaDesc: "AI content",
		CrawledAt: time.Now(),
		IndexedAt: time.Now(),
		StatusCode: 200,
		Hash:      "aihash123",
	}
	if err := db.SaveDocument(doc); err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}
	
	// Semantic search
	reqBody := map[string]interface{}{
		"query": "artificial intelligence",
		"limit": 10,
	}
	body, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/api/semantic-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["type"] != "semantic" {
		t.Errorf("Expected type 'semantic', got '%v'", response["type"])
	}
	
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		if metadata["search_type"] != "semantic" {
			t.Errorf("Expected search_type 'semantic', got '%v'", metadata["search_type"])
		}
	} else {
		t.Error("Expected metadata in response")
	}
}
