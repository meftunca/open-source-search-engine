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

package database

import (
	"os"
	"testing"
	"time"

	"github.com/meftunca/open-source-search-engine/internal/models"
)

func setupTestDB(t *testing.T) *DB {
	tmpFile := "/tmp/test_db_" + time.Now().Format("20060102150405") + ".db"
	db, err := New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})
	
	return db
}

func TestSaveAndGetDocument(t *testing.T) {
	db := setupTestDB(t)
	
	doc := &models.Document{
		URL:      "https://example.com",
		Title:    "Example Domain",
		Content:  "This domain is for use in illustrative examples",
		MetaDesc: "Example domain description",
		Keywords: "example, domain",
		ImageURLs: `["https://example.com/img1.jpg"]`,
		VideoURLs: `["https://example.com/video1.mp4"]`,
		OGImage:  "https://example.com/og-image.jpg",
		OGTitle:  "Example OG Title",
		OGDesc:   "Example OG Description",
		Author:   "Test Author",
		CrawledAt: time.Now(),
		IndexedAt: time.Now(),
		StatusCode: 200,
		ContentType: "text/html",
		Hash:      "abc123",
	}
	
	if err := db.SaveDocument(doc); err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	
	retrieved, err := db.GetDocument(doc.URL)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	
	if retrieved.URL != doc.URL {
		t.Errorf("Expected URL '%s', got '%s'", doc.URL, retrieved.URL)
	}
	
	if retrieved.Title != doc.Title {
		t.Errorf("Expected title '%s', got '%s'", doc.Title, retrieved.Title)
	}
	
	if retrieved.ImageURLs != doc.ImageURLs {
		t.Errorf("Expected ImageURLs '%s', got '%s'", doc.ImageURLs, retrieved.ImageURLs)
	}
	
	if retrieved.OGTitle != doc.OGTitle {
		t.Errorf("Expected OGTitle '%s', got '%s'", doc.OGTitle, retrieved.OGTitle)
	}
}

func TestSearch(t *testing.T) {
	db := setupTestDB(t)
	
	// Add test documents
	docs := []*models.Document{
		{
			URL:      "https://golang.org",
			Title:    "The Go Programming Language",
			Content:  "Go is an open source programming language",
			MetaDesc: "Go programming language",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			Hash:      "hash1",
		},
		{
			URL:      "https://python.org",
			Title:    "Python Programming",
			Content:  "Python is a high-level programming language",
			MetaDesc: "Python language",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			Hash:      "hash2",
		},
	}
	
	for _, doc := range docs {
		if err := db.SaveDocument(doc); err != nil {
			t.Fatalf("Failed to save document: %v", err)
		}
	}
	
	// Search for "programming"
	results, err := db.Search("programming", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}
	
	// Search for "Go"
	results, err = db.Search("Go", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) < 1 {
		t.Errorf("Expected at least 1 result, got %d", len(results))
	}
}

func TestSearchImages(t *testing.T) {
	db := setupTestDB(t)
	
	// Add documents with and without images
	docs := []*models.Document{
		{
			URL:       "https://example.com/with-images",
			Title:     "Page with Images",
			Content:   "This page has images",
			ImageURLs: `["https://example.com/img1.jpg", "https://example.com/img2.jpg"]`,
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			Hash:      "hash_img1",
		},
		{
			URL:       "https://example.com/no-images",
			Title:     "Page without Images",
			Content:   "This page has no images",
			ImageURLs: "",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			Hash:      "hash_img2",
		},
	}
	
	for _, doc := range docs {
		if err := db.SaveDocument(doc); err != nil {
			t.Fatalf("Failed to save document: %v", err)
		}
	}
	
	// Search for images
	results, err := db.SearchImages("page", 10)
	if err != nil {
		t.Fatalf("Image search failed: %v", err)
	}
	
	// Should only return the document with images
	if len(results) != 1 {
		t.Errorf("Expected 1 result with images, got %d", len(results))
	}
	
	if len(results) > 0 && results[0].ImageURLs == "" {
		t.Error("Expected result to have images")
	}
}

func TestSearchVideos(t *testing.T) {
	db := setupTestDB(t)
	
	// Add documents with and without videos
	docs := []*models.Document{
		{
			URL:       "https://example.com/with-videos",
			Title:     "Page with Videos",
			Content:   "This page has videos",
			VideoURLs: `["https://youtube.com/watch?v=123"]`,
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			Hash:      "hash_vid1",
		},
		{
			URL:       "https://example.com/no-videos",
			Title:     "Page without Videos",
			Content:   "This page has no videos",
			VideoURLs: "",
			CrawledAt: time.Now(),
			IndexedAt: time.Now(),
			StatusCode: 200,
			Hash:      "hash_vid2",
		},
	}
	
	for _, doc := range docs {
		if err := db.SaveDocument(doc); err != nil {
			t.Fatalf("Failed to save document: %v", err)
		}
	}
	
	// Search for videos
	results, err := db.SearchVideos("page", 10)
	if err != nil {
		t.Fatalf("Video search failed: %v", err)
	}
	
	// Should only return the document with videos
	if len(results) != 1 {
		t.Errorf("Expected 1 result with videos, got %d", len(results))
	}
	
	if len(results) > 0 && results[0].VideoURLs == "" {
		t.Error("Expected result to have videos")
	}
}

func TestAddURL(t *testing.T) {
	db := setupTestDB(t)
	
	err := db.AddURL("https://example.com", 10)
	if err != nil {
		t.Fatalf("Failed to add URL: %v", err)
	}
	
	// Try adding the same URL again (should not error due to UNIQUE constraint handling)
	err = db.AddURL("https://example.com", 10)
	if err != nil {
		t.Fatalf("Failed to add duplicate URL: %v", err)
	}
}

func TestGetNextURLs(t *testing.T) {
	db := setupTestDB(t)
	
	// Add URLs with different priorities
	urls := []struct {
		url      string
		priority int
	}{
		{"https://example.com/high", 100},
		{"https://example.com/medium", 50},
		{"https://example.com/low", 10},
	}
	
	for _, u := range urls {
		if err := db.AddURL(u.url, u.priority); err != nil {
			t.Fatalf("Failed to add URL: %v", err)
		}
	}
	
	// Get next URLs (should return highest priority first)
	nextURLs, err := db.GetNextURLs(2)
	if err != nil {
		t.Fatalf("Failed to get next URLs: %v", err)
	}
	
	if len(nextURLs) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(nextURLs))
	}
	
	// First URL should have highest priority
	if len(nextURLs) > 0 && nextURLs[0].Priority != 100 {
		t.Errorf("Expected first URL to have priority 100, got %d", nextURLs[0].Priority)
	}
}

func TestURLQueueOperations(t *testing.T) {
	db := setupTestDB(t)
	
	// Test adding URLs
	err := db.AddURL("https://test1.com", 100)
	if err != nil {
		t.Fatalf("Failed to add URL: %v", err)
	}
	
	err = db.AddURL("https://test2.com", 50)
	if err != nil {
		t.Fatalf("Failed to add URL: %v", err)
	}
	
	// Test getting URLs - should get highest priority first
	urls, err := db.GetNextURLs(2)
	if err != nil {
		t.Fatalf("Failed to get URLs: %v", err)
	}
	
	if len(urls) != 2 {
		t.Fatalf("Expected 2 URLs, got %d", len(urls))
	}
	
	if urls[0].URL != "https://test1.com" {
		t.Errorf("Expected first URL to be test1.com (highest priority), got %s", urls[0].URL)
	}
	
	if urls[1].URL != "https://test2.com" {
		t.Errorf("Expected second URL to be test2.com, got %s", urls[1].URL)
	}
	
	// Verify priorities
	if urls[0].Priority != 100 {
		t.Errorf("Expected first URL priority 100, got %d", urls[0].Priority)
	}
}
