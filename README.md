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
- 💾 **DuckDB Storage** - Efficient analytical database for indexing
- 🌐 **REST API** - Simple HTTP API for search and crawl management
- 📊 **Web Interface** - Built-in web UI for searching
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
go test ./...
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

1. **Crawler**: Fetches web pages concurrently using goroutines, respecting crawl delays and politeness policies.

2. **Database**: Uses DuckDB for efficient storage and querying of documents, URLs, and metadata.

3. **Indexer**: Extracts and indexes content from crawled pages, including titles, meta descriptions, and full text.

4. **Search API**: Provides RESTful endpoints for searching and managing the crawl queue.

5. **Web Interface**: Simple HTML/JavaScript frontend for interactive searching.

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
