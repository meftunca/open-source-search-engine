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

package models

import "time"

// Document represents a crawled and indexed web page
type Document struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	MetaDesc    string    `json:"meta_description"`
	Keywords    string    `json:"keywords"`
	CrawledAt   time.Time `json:"crawled_at"`
	IndexedAt   time.Time `json:"indexed_at"`
	StatusCode  int       `json:"status_code"`
	ContentType string    `json:"content_type"`
	Hash        string    `json:"hash"`
}

// URLQueue represents a URL waiting to be crawled
type URLQueue struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Priority    int       `json:"priority"`
	AddedAt     time.Time `json:"added_at"`
	LastAttempt *time.Time `json:"last_attempt,omitempty"`
	Attempts    int       `json:"attempts"`
	Status      string    `json:"status"` // pending, processing, completed, failed
}

// SearchResult represents a search result
type SearchResult struct {
	Document
	Score float64 `json:"score"`
	Snippet string `json:"snippet"`
}
