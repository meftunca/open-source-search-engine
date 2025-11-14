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
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/meftunca/open-source-search-engine/pkg/database"
)

// Server represents the HTTP API server
type Server struct {
	db   *database.DB
	mux  *http.ServeMux
}

// New creates a new API server
func New(db *database.DB) *Server {
	s := &Server{
		db:  db,
		mux: http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// registerRoutes sets up the API endpoints
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/crawl", s.handleCrawl)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/", s.handleRoot)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleSearch handles search requests
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := s.db.Search(query, limit)
	if err != nil {
		log.Printf("Search error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate snippets
	for _, result := range results {
		result.Snippet = generateSnippet(result.Content, query)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

// handleCrawl handles requests to add URLs to the crawl queue
func (s *Server) handleCrawl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL      string `json:"url"`
		Priority int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if err := s.db.AddURL(req.URL, req.Priority); err != nil {
		log.Printf("Error adding URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "URL added to crawl queue",
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleRoot serves a simple welcome page
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Open Source Search Engine</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .search-box { margin: 20px 0; }
        input[type="text"] { width: 60%; padding: 10px; font-size: 16px; }
        button { padding: 10px 20px; font-size: 16px; cursor: pointer; }
        .results { margin-top: 20px; }
        .result { margin: 15px 0; padding: 10px; border-left: 3px solid #4CAF50; }
        .result h3 { margin: 0 0 5px 0; }
        .result .url { color: #666; font-size: 14px; }
        .result .snippet { margin-top: 5px; }
    </style>
</head>
<body>
    <h1>Open Source Search Engine</h1>
    <div class="search-box">
        <input type="text" id="query" placeholder="Enter search query...">
        <button onclick="search()">Search</button>
    </div>
    <div id="results" class="results"></div>

    <h2>API Endpoints</h2>
    <ul>
        <li><strong>GET /api/search?q=QUERY</strong> - Search for documents</li>
        <li><strong>POST /api/crawl</strong> - Add URL to crawl queue (JSON: {"url": "...", "priority": 0})</li>
        <li><strong>GET /api/health</strong> - Health check</li>
    </ul>

    <script>
        function search() {
            const query = document.getElementById('query').value;
            if (!query) return;

            fetch('/api/search?q=' + encodeURIComponent(query))
                .then(r => r.json())
                .then(data => {
                    const resultsDiv = document.getElementById('results');
                    if (data.count === 0) {
                        resultsDiv.innerHTML = '<p>No results found.</p>';
                        return;
                    }
                    
                    let html = '<h3>Results (' + data.count + ')</h3>';
                    data.results.forEach(r => {
                        html += '<div class="result">';
                        html += '<h3>' + escapeHtml(r.title || 'Untitled') + '</h3>';
                        html += '<div class="url">' + escapeHtml(r.url) + '</div>';
                        html += '<div class="snippet">' + escapeHtml(r.snippet || r.meta_description || '') + '</div>';
                        html += '</div>';
                    });
                    resultsDiv.innerHTML = html;
                })
                .catch(err => {
                    console.error('Search error:', err);
                    document.getElementById('results').innerHTML = '<p>Error performing search.</p>';
                });
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        document.getElementById('query').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') search();
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// generateSnippet creates a snippet of text around the query match
func generateSnippet(content, query string) string {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return ""
	}

	// Simple snippet generation - find query in content
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)
	
	index := strings.Index(lowerContent, lowerQuery)
	if index == -1 {
		// Query not found, return first 200 chars
		if len(content) > 200 {
			return content[:200] + "..."
		}
		return content
	}

	// Extract context around match
	start := index - 100
	if start < 0 {
		start = 0
	}
	end := index + len(query) + 100
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}
