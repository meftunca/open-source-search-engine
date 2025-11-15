# Open Source Search Engine

A distributed web search engine and crawler built with Go and DuckDB.

## Overview

This is a modern, high-performance search engine that leverages:
- **Go** for concurrent crawling and efficient API serving
- **DuckDB** for fast analytical queries and document storage
- **goroutines** for parallel web crawling
- **RESTful API** for search and management

Originally based on Gigablast (http://www.gigablast.com/), this project has been completely reorganized and rewritten in Go for better performance, maintainability, and modern architecture.

## Features

- 🚀 **Concurrent Web Crawler** - Parallel crawling using Go goroutines
- 🔍 **Full-Text Search** - Fast search across crawled documents
- 🖼️ **Image Search** - Dedicated endpoint for searching documents with images
- 🎥 **Video Search** - Dedicated endpoint for searching documents with videos
- 🤖 **Semantic Search API** - AI-ready endpoint for semantic search and LLM integration
- 💾 **DuckDB Storage** - Efficient analytical database for indexing
- 🌐 **REST API** - Simple HTTP API for search and crawl management
- 📊 **Web Interface** - Built-in web UI for searching
- 📝 **Enhanced Metadata** - Extract Open Graph, keywords, author, images, and videos
- 📈 **Monitoring** - Built-in metrics and structured logging with logrus
- ⚙️ **Configurable** - Environment-based configuration

## Requirements

- Go 1.20 or higher
- Linux (Intel/AMD architecture)
- Internet connection for crawling

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/meftunca/open-source-search-engine.git
cd open-source-search-engine

# Download dependencies
go mod download

# Build the application
go build -o search-engine ./cmd/search-engine

# Run the search engine
./search-engine
```

### Quick Start

```bash
# Run with default configuration
go run ./cmd/search-engine/main.go
```

The server will start on `http://localhost:8080`

## Configuration

Configure the search engine using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_PATH` | `./search_engine.db` | Path to DuckDB database file |
| `HTTP_PORT` | `8080` | HTTP server port |
| `CRAWLER_WORKERS` | `1` | Number of concurrent crawler workers (Note: DuckDB has limited concurrent write support, so 1 worker is recommended) |
| `CRAWL_DELAY` | `1` | Delay in seconds between requests to same domain |
| `MAX_DEPTH` | `3` | Maximum crawl depth |
| `USER_AGENT` | `OpenSourceSearchEngine/1.0` | User agent string |
| `ROBOTS_ENABLED` | `true` | Respect robots.txt |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |

### Example

```bash
export DB_PATH="./my_index.db"
export HTTP_PORT="9090"
export CRAWLER_WORKERS="20"
go run ./cmd/search-engine/main.go
```

## API Endpoints

### Search

```bash
GET /api/search?q=<query>&limit=<limit>
```

Search for documents matching the query.

**Parameters:**
- `q` (required): Search query
- `limit` (optional): Maximum number of results (default: 10)

**Example:**
```bash
curl "http://localhost:8080/api/search?q=golang&limit=5"
```

### Image Search

```bash
GET /api/search/images?q=<query>&limit=<limit>
```

Search for documents containing images.

**Parameters:**
- `q` (required): Search query
- `limit` (optional): Maximum number of results (default: 10)

**Example:**
```bash
curl "http://localhost:8080/api/search/images?q=tutorial&limit=5"
```

### Video Search

```bash
GET /api/search/videos?q=<query>&limit=<limit>
```

Search for documents containing videos.

**Parameters:**
- `q` (required): Search query
- `limit` (optional): Maximum number of results (default: 10)

**Example:**
```bash
curl "http://localhost:8080/api/search/videos?q=programming&limit=5"
```

### Semantic Search (AI Integration)

```bash
POST /api/semantic-search
Content-Type: application/json

{
  "query": "your search query",
  "limit": 10,
  "embedding": [0.1, 0.2, ...],  // optional: for future AI integration
  "metadata": {}                   // optional: additional context
}
```

Semantic search endpoint designed for AI and LLM integration. Returns enhanced metadata suitable for AI consumption.

**Example:**
```bash
curl -X POST http://localhost:8080/api/semantic-search \
  -H "Content-Type: application/json" \
  -d '{"query": "artificial intelligence", "limit": 5}'
```

**Response includes:**
- Enhanced document metadata (Open Graph, author, keywords)
- Image and video URLs
- Structured data for AI processing

### Add URL to Crawl Queue

```bash
POST /api/crawl
Content-Type: application/json

{
  "url": "https://example.com",
  "priority": 10
}
```

Add a URL to the crawl queue.

**Example:**
```bash
curl -X POST http://localhost:8080/api/crawl \
  -H "Content-Type: application/json" \
  -d '{"url": "https://go.dev", "priority": 10}'
```

### Health Check

```bash
GET /api/health
```

Check if the service is running.

### Metrics

```bash
GET /api/metrics
```

Get system metrics including request counts, crawl statistics, and performance data.

**Example:**
```bash
curl "http://localhost:8080/api/metrics"
```

**Response includes:**
- `total_requests`: Total number of API requests
- `search_requests`: Number of search requests
- `image_searches`: Number of image search requests
- `video_searches`: Number of video search requests
- `semantic_searches`: Number of semantic search requests
- `pages_crawled`: Total pages crawled
- `crawl_errors`: Number of crawl errors
- `avg_response_time_ms`: Average API response time
- `uptime_seconds`: Server uptime

## Usage

1. **Start the search engine:**
   ```bash
   ./search-engine
   ```

2. **Add URLs to crawl:**
   ```bash
   curl -X POST http://localhost:8080/api/crawl \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com", "priority": 10}'
   ```

3. **Search for content:**
   ```bash
   curl "http://localhost:8080/api/search?q=your+search+query"
   ```

4. **Use the web interface:**
   Open your browser and navigate to `http://localhost:8080`

## Project Structure

```
.
├── cmd/
│   └── search-engine/     # Main application entry point
│       └── main.go
├── pkg/
│   ├── api/              # HTTP API server
│   ├── config/           # Configuration management
│   ├── crawler/          # Web crawler implementation
│   ├── database/         # DuckDB integration
│   └── search/           # Search functionality
├── internal/
│   └── models/           # Data models
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── README.md             # This file
└── LICENSE               # Apache 2.0 License
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/api -v
go test ./pkg/crawler -v
go test ./pkg/database -v
```

### Logging

The search engine uses structured JSON logging via logrus. Logs include:
- HTTP request details (method, path, duration, remote IP)
- Crawler operations (URLs being crawled, success/failure)
- Database operations
- Error tracking

Set `LOG_LEVEL` environment variable to control verbosity:
```bash
export LOG_LEVEL=debug  # Options: debug, info, warn, error
```

### Building for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o search-engine ./cmd/search-engine

# Or build with version info
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.Version=$VERSION" -o search-engine ./cmd/search-engine
```

## Architecture

The search engine consists of several key components:

1. **Crawler**: Fetches web pages concurrently using goroutines, respecting crawl delays and politeness policies. Enhanced with:
   - Comprehensive metadata extraction (Open Graph, Twitter Cards, keywords, author)
   - Image and video URL extraction
   - HTML cleaning (removes scripts and styles)
   - Structured data extraction

2. **Database**: Uses DuckDB for efficient storage and querying of documents, URLs, and metadata. Supports:
   - Full-text search
   - Image-specific search
   - Video-specific search
   - Priority-based URL queue management

3. **Indexer**: Extracts and indexes content from crawled pages, including:
   - Titles and meta descriptions
   - Full text content
   - Keywords and author information
   - Open Graph metadata
   - Image and video URLs

4. **Search API**: Provides RESTful endpoints for:
   - Standard text search
   - Image search
   - Video search
   - Semantic search (AI-ready)
   - Crawl queue management
   - System metrics

5. **Web Interface**: Simple HTML/JavaScript frontend for interactive searching.

6. **Logging & Monitoring**: Structured JSON logging with logrus and basic metrics tracking.

## AI Integration

The semantic search endpoint (`/api/semantic-search`) is designed for AI and LLM integration:

- **Enhanced Metadata**: Returns comprehensive metadata including Open Graph tags, author, keywords
- **Media URLs**: Provides arrays of image and video URLs for multimodal AI
- **Structured Response**: JSON format optimized for AI consumption
- **Future-Ready**: Placeholder for embedding-based semantic search

Example AI integration:
```python
import requests

response = requests.post('http://localhost:8080/api/semantic-search', json={
    'query': 'machine learning tutorials',
    'limit': 10
})

results = response.json()
for doc in results['results']:
    print(f"Title: {doc['title']}")
    print(f"URL: {doc['url']}")
    print(f"Images: {doc['image_urls']}")
    print(f"Videos: {doc['video_urls']}")
```

## Performance

The Go implementation provides significant performance improvements over the original C++ version:
- **Concurrency**: Native goroutines for efficient parallel processing
- **Memory Safety**: Go's garbage collector eliminates memory management issues
- **Fast Compilation**: Quick build times for rapid development
- **DuckDB**: Columnar storage for fast analytical queries

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Original Copyright 2013 Web Research Properties, LLC and Matt Wells and Gigablast, Inc.  
Go Implementation Copyright 2025 Open Source Search Engine Contributors

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Contact

For questions, feature requests, or support, please open an issue on GitHub.

## Acknowledgments

This project is based on the original Gigablast search engine by Matt Wells, completely reorganized and rewritten in Go with modern architecture and technologies.
