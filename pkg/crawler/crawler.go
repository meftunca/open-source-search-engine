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

package crawler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/meftunca/open-source-search-engine/internal/models"
	"github.com/meftunca/open-source-search-engine/pkg/config"
	"github.com/meftunca/open-source-search-engine/pkg/database"
	"golang.org/x/net/html"
)

// Crawler manages the web crawling process
type Crawler struct {
	db     *database.DB
	config *config.Config
	client *http.Client
	wg     sync.WaitGroup
}

// New creates a new crawler instance
func New(db *database.DB, cfg *config.Config) *Crawler {
	return &Crawler{
		db:     db,
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Start begins the crawling process
func (c *Crawler) Start(ctx context.Context) error {
	log.Println("Starting crawler with", c.config.CrawlerWorkers, "workers")

	for i := 0; i < c.config.CrawlerWorkers; i++ {
		c.wg.Add(1)
		go c.worker(ctx, i)
	}

	c.wg.Wait()
	return nil
}

// worker processes URLs from the queue
func (c *Crawler) worker(ctx context.Context, id int) {
	defer c.wg.Done()

	log.Printf("Worker %d started\n", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping\n", id)
			return
		default:
			// Get next URL from queue
			urls, err := c.db.GetNextURLs(1)
			if err != nil {
				log.Printf("Worker %d: error getting URLs: %v\n", id, err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(urls) == 0 {
				// No URLs to process, sleep and retry
				time.Sleep(2 * time.Second)
				continue
			}

			urlItem := urls[0]
			log.Printf("Worker %d: crawling %s\n", id, urlItem.URL)

			// Update status to processing
			if err := c.db.UpdateURLStatus(urlItem.ID, "processing"); err != nil {
				log.Printf("Worker %d: error updating status: %v\n", id, err)
				continue
			}

			// Crawl the URL
			doc, err := c.crawl(urlItem.URL)
			if err != nil {
				log.Printf("Worker %d: error crawling %s: %v\n", id, urlItem.URL, err)
				c.db.UpdateURLStatus(urlItem.ID, "failed")
				continue
			}

			// Save the document
			if err := c.db.SaveDocument(doc); err != nil {
				log.Printf("Worker %d: error saving document: %v\n", id, err)
				c.db.UpdateURLStatus(urlItem.ID, "failed")
				continue
			}

			// Mark as completed
			c.db.UpdateURLStatus(urlItem.ID, "completed")

			// Extract and queue new URLs
			c.extractAndQueueURLs(doc.Content, urlItem.URL)

			// Respect crawl delay
			time.Sleep(time.Duration(c.config.CrawlDelay) * time.Second)
		}
	}
}

// crawl fetches and parses a URL
func (c *Crawler) crawl(urlStr string) (*models.Document, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	title := extractTitle(doc)
	metaDesc := extractMetaDescription(doc)
	content := extractText(doc)

	// Calculate hash
	hash := fmt.Sprintf("%x", sha256.Sum256(body))

	return &models.Document{
		URL:         urlStr,
		Title:       title,
		Content:     content,
		MetaDesc:    metaDesc,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		CrawledAt:   time.Now(),
		IndexedAt:   time.Now(),
		Hash:        hash,
	}, nil
}

// extractTitle extracts the title from HTML
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return n.FirstChild.Data
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}
	return ""
}

// extractMetaDescription extracts meta description from HTML
func extractMetaDescription(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var name, content string
		for _, attr := range n.Attr {
			if attr.Key == "name" && attr.Val == "description" {
				name = attr.Val
			}
			if attr.Key == "content" {
				content = attr.Val
			}
		}
		if name == "description" && content != "" {
			return content
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if desc := extractMetaDescription(c); desc != "" {
			return desc
		}
	}
	return ""
}

// extractText extracts visible text from HTML
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += extractText(c) + " "
	}
	return strings.TrimSpace(text)
}

// extractAndQueueURLs extracts URLs from content and adds them to queue
func (c *Crawler) extractAndQueueURLs(content, baseURL string) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return
	}

	var extractURLs func(*html.Node)
	extractURLs = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					if absoluteURL, err := toAbsoluteURL(href, baseURL); err == nil {
						// Add to queue with lower priority
						c.db.AddURL(absoluteURL, 0)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractURLs(c)
		}
	}

	extractURLs(doc)
}

// toAbsoluteURL converts relative URLs to absolute
func toAbsoluteURL(href, baseURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	absoluteURL := base.ResolveReference(hrefURL)
	
	// Only HTTP(S) URLs
	if absoluteURL.Scheme != "http" && absoluteURL.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", absoluteURL.Scheme)
	}

	return absoluteURL.String(), nil
}
