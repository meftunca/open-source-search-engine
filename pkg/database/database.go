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
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/meftunca/open-source-search-engine/internal/models"
)

// DB wraps the DuckDB connection
type DB struct {
	conn *sql.DB
	mu   sync.Mutex // Mutex for concurrent access to queue operations
}

// New creates a new database connection and initializes tables
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initTables creates the necessary database tables
func (db *DB) initTables() error {
	queries := []string{
		`CREATE SEQUENCE IF NOT EXISTS seq_documents`,
		`CREATE SEQUENCE IF NOT EXISTS seq_url_queue`,
		`CREATE TABLE IF NOT EXISTS documents (
			id INTEGER DEFAULT nextval('seq_documents'),
			url VARCHAR NOT NULL UNIQUE,
			title VARCHAR,
			content TEXT,
			meta_description VARCHAR,
			keywords VARCHAR,
			image_urls TEXT,
			video_urls TEXT,
			og_image VARCHAR,
			og_title VARCHAR,
			og_description VARCHAR,
			author VARCHAR,
			crawled_at TIMESTAMP,
			indexed_at TIMESTAMP,
			status_code INTEGER,
			content_type VARCHAR,
			hash VARCHAR,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS url_queue (
			id INTEGER DEFAULT nextval('seq_url_queue'),
			url VARCHAR NOT NULL UNIQUE,
			priority INTEGER DEFAULT 0,
			added_at TIMESTAMP,
			last_attempt TIMESTAMP,
			attempts INTEGER DEFAULT 0,
			status VARCHAR DEFAULT 'pending',
			PRIMARY KEY (id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_documents_url ON documents(url)`,
		`CREATE INDEX IF NOT EXISTS idx_url_queue_status ON url_queue(status)`,
		`CREATE INDEX IF NOT EXISTS idx_url_queue_priority ON url_queue(priority DESC)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// SaveDocument saves or updates a document in the database
func (db *DB) SaveDocument(doc *models.Document) error {
	query := `INSERT INTO documents 
		(url, title, content, meta_description, keywords, image_urls, video_urls, og_image, og_title, og_description, author, crawled_at, indexed_at, status_code, content_type, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (url) DO UPDATE SET
		title = EXCLUDED.title,
		content = EXCLUDED.content,
		meta_description = EXCLUDED.meta_description,
		keywords = EXCLUDED.keywords,
		image_urls = EXCLUDED.image_urls,
		video_urls = EXCLUDED.video_urls,
		og_image = EXCLUDED.og_image,
		og_title = EXCLUDED.og_title,
		og_description = EXCLUDED.og_description,
		author = EXCLUDED.author,
		crawled_at = EXCLUDED.crawled_at,
		indexed_at = EXCLUDED.indexed_at,
		status_code = EXCLUDED.status_code,
		content_type = EXCLUDED.content_type,
		hash = EXCLUDED.hash`

	_, err := db.conn.Exec(query,
		doc.URL, doc.Title, doc.Content, doc.MetaDesc, doc.Keywords,
		doc.ImageURLs, doc.VideoURLs, doc.OGImage, doc.OGTitle, doc.OGDesc, doc.Author,
		doc.CrawledAt, doc.IndexedAt, doc.StatusCode, doc.ContentType, doc.Hash)

	return err
}

// GetDocument retrieves a document by URL
func (db *DB) GetDocument(url string) (*models.Document, error) {
	query := `SELECT id, url, title, content, meta_description, keywords, 
		image_urls, video_urls, og_image, og_title, og_description, author,
		crawled_at, indexed_at, status_code, content_type, hash
		FROM documents WHERE url = ?`

	var doc models.Document
	err := db.conn.QueryRow(query, url).Scan(
		&doc.ID, &doc.URL, &doc.Title, &doc.Content, &doc.MetaDesc, &doc.Keywords,
		&doc.ImageURLs, &doc.VideoURLs, &doc.OGImage, &doc.OGTitle, &doc.OGDesc, &doc.Author,
		&doc.CrawledAt, &doc.IndexedAt, &doc.StatusCode, &doc.ContentType, &doc.Hash)

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

// Search performs a full-text search on documents
func (db *DB) Search(query string, limit int) ([]*models.SearchResult, error) {
	// Simple search implementation using LIKE for now
	// In production, you'd want to use FTS or more advanced indexing
	searchQuery := `SELECT id, url, title, content, meta_description, keywords,
		image_urls, video_urls, og_image, og_title, og_description, author,
		crawled_at, indexed_at, status_code, content_type, hash
		FROM documents
		WHERE title LIKE ? OR content LIKE ? OR meta_description LIKE ?
		LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := db.conn.Query(searchQuery, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		err := rows.Scan(
			&result.ID, &result.URL, &result.Title, &result.Content,
			&result.MetaDesc, &result.Keywords,
			&result.ImageURLs, &result.VideoURLs, &result.OGImage, &result.OGTitle, &result.OGDesc, &result.Author,
			&result.CrawledAt,
			&result.IndexedAt, &result.StatusCode, &result.ContentType, &result.Hash)
		if err != nil {
			return nil, err
		}
		result.Score = 1.0 // Simple scoring for now
		results = append(results, &result)
	}

	return results, nil
}

// AddURL adds a URL to the crawl queue
func (db *DB) AddURL(url string, priority int) error {
	query := `INSERT INTO url_queue (url, priority, added_at, status)
		VALUES (?, ?, CURRENT_TIMESTAMP, 'pending')
		ON CONFLICT (url) DO NOTHING`

	_, err := db.conn.Exec(query, url, priority)
	return err
}

// GetNextURLs retrieves the next URLs to crawl
func (db *DB) GetNextURLs(limit int) ([]*models.URLQueue, error) {
	// Use mutex to prevent concurrent access to URL queue operations
	db.mu.Lock()
	defer db.mu.Unlock()

	query := `SELECT id, url, priority, added_at, last_attempt, attempts, status
		FROM url_queue
		WHERE status = 'pending'
		ORDER BY priority DESC, added_at ASC
		LIMIT ?`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*models.URLQueue
	for rows.Next() {
		var u models.URLQueue
		err := rows.Scan(&u.ID, &u.URL, &u.Priority, &u.AddedAt, &u.LastAttempt, &u.Attempts, &u.Status)
		if err != nil {
			return nil, err
		}
		urls = append(urls, &u)
	}

	return urls, nil
}

// UpdateURLStatus updates the status of a URL in the queue
func (db *DB) UpdateURLStatus(id int64, status string) error {
	query := `UPDATE url_queue SET status = ? WHERE id = ?`
	_, err := db.conn.Exec(query, status, id)
	return err
}

// SearchImages searches for documents that contain images
func (db *DB) SearchImages(query string, limit int) ([]*models.SearchResult, error) {
	searchQuery := `SELECT id, url, title, content, meta_description, keywords,
		image_urls, video_urls, og_image, og_title, og_description, author,
		crawled_at, indexed_at, status_code, content_type, hash
		FROM documents
		WHERE (title LIKE ? OR content LIKE ? OR meta_description LIKE ?)
		AND (image_urls IS NOT NULL AND image_urls != '' AND image_urls != '[]')
		LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := db.conn.Query(searchQuery, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		err := rows.Scan(
			&result.ID, &result.URL, &result.Title, &result.Content,
			&result.MetaDesc, &result.Keywords,
			&result.ImageURLs, &result.VideoURLs, &result.OGImage, &result.OGTitle, &result.OGDesc, &result.Author,
			&result.CrawledAt,
			&result.IndexedAt, &result.StatusCode, &result.ContentType, &result.Hash)
		if err != nil {
			return nil, err
		}
		result.Score = 1.0
		results = append(results, &result)
	}

	return results, nil
}

// SearchVideos searches for documents that contain videos
func (db *DB) SearchVideos(query string, limit int) ([]*models.SearchResult, error) {
	searchQuery := `SELECT id, url, title, content, meta_description, keywords,
		image_urls, video_urls, og_image, og_title, og_description, author,
		crawled_at, indexed_at, status_code, content_type, hash
		FROM documents
		WHERE (title LIKE ? OR content LIKE ? OR meta_description LIKE ?)
		AND (video_urls IS NOT NULL AND video_urls != '' AND video_urls != '[]')
		LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := db.conn.Query(searchQuery, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		err := rows.Scan(
			&result.ID, &result.URL, &result.Title, &result.Content,
			&result.MetaDesc, &result.Keywords,
			&result.ImageURLs, &result.VideoURLs, &result.OGImage, &result.OGTitle, &result.OGDesc, &result.Author,
			&result.CrawledAt,
			&result.IndexedAt, &result.StatusCode, &result.ContentType, &result.Hash)
		if err != nil {
			return nil, err
		}
		result.Score = 1.0
		results = append(results, &result)
	}

	return results, nil
}
