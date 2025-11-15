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

package metrics

import (
	"sync"
	"time"
)

// Metrics holds basic system metrics
type Metrics struct {
	mu sync.RWMutex
	
	// API metrics
	TotalRequests     int64
	SearchRequests    int64
	ImageSearches     int64
	VideoSearches     int64
	SemanticSearches  int64
	
	// Crawler metrics
	PagesCrawled      int64
	CrawlErrors       int64
	
	// Response times
	AvgResponseTime   time.Duration
	
	StartTime         time.Time
}

var globalMetrics *Metrics
var once sync.Once

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		globalMetrics = &Metrics{
			StartTime: time.Now(),
		}
	})
	return globalMetrics
}

// IncrementRequests increments total request count
func (m *Metrics) IncrementRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
}

// IncrementSearchRequests increments search request count
func (m *Metrics) IncrementSearchRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SearchRequests++
}

// IncrementImageSearches increments image search count
func (m *Metrics) IncrementImageSearches() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ImageSearches++
}

// IncrementVideoSearches increments video search count
func (m *Metrics) IncrementVideoSearches() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.VideoSearches++
}

// IncrementSemanticSearches increments semantic search count
func (m *Metrics) IncrementSemanticSearches() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SemanticSearches++
}

// IncrementPagesCrawled increments pages crawled count
func (m *Metrics) IncrementPagesCrawled() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PagesCrawled++
}

// IncrementCrawlErrors increments crawl errors count
func (m *Metrics) IncrementCrawlErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CrawlErrors++
}

// UpdateAvgResponseTime updates the average response time
func (m *Metrics) UpdateAvgResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.AvgResponseTime == 0 {
		m.AvgResponseTime = duration
	} else {
		m.AvgResponseTime = (m.AvgResponseTime + duration) / 2
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{
		"total_requests":      m.TotalRequests,
		"search_requests":     m.SearchRequests,
		"image_searches":      m.ImageSearches,
		"video_searches":      m.VideoSearches,
		"semantic_searches":   m.SemanticSearches,
		"pages_crawled":       m.PagesCrawled,
		"crawl_errors":        m.CrawlErrors,
		"avg_response_time_ms": m.AvgResponseTime.Milliseconds(),
		"uptime_seconds":      time.Since(m.StartTime).Seconds(),
	}
}
