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

package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	DatabasePath    string
	HTTPPort        string
	CrawlerWorkers  int
	CrawlDelay      int // seconds between requests to the same domain
	MaxDepth        int
	UserAgent       string
	RobotsEnabled   bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		DatabasePath:   getEnv("DB_PATH", "./search_engine.db"),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		CrawlerWorkers: getEnvAsInt("CRAWLER_WORKERS", 1),
		CrawlDelay:     getEnvAsInt("CRAWL_DELAY", 1),
		MaxDepth:       getEnvAsInt("MAX_DEPTH", 3),
		UserAgent:      getEnv("USER_AGENT", "OpenSourceSearchEngine/1.0"),
		RobotsEnabled:  getEnvAsBool("ROBOTS_ENABLED", true),
	}
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}
