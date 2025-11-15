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
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
)

// extractEnhancedMetadata extracts comprehensive metadata from HTML
func extractEnhancedMetadata(doc *html.Node, baseURL string) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	metadata["title"] = extractTitle(doc)
	metadata["meta_description"] = extractMetaDescription(doc)
	metadata["keywords"] = extractKeywords(doc)
	metadata["author"] = extractAuthor(doc)
	metadata["og_title"] = extractOpenGraphTag(doc, "og:title")
	metadata["og_description"] = extractOpenGraphTag(doc, "og:description")
	metadata["og_image"] = extractOpenGraphTag(doc, "og:image")
	
	// Extract images
	images := extractImages(doc, baseURL)
	imagesJSON, _ := json.Marshal(images)
	metadata["image_urls"] = string(imagesJSON)
	
	// Extract videos
	videos := extractVideos(doc, baseURL)
	videosJSON, _ := json.Marshal(videos)
	metadata["video_urls"] = string(videosJSON)
	
	return metadata
}

// extractKeywords extracts keywords from meta tags
func extractKeywords(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var name, content string
		for _, attr := range n.Attr {
			if attr.Key == "name" && (attr.Val == "keywords" || attr.Val == "Keywords") {
				name = attr.Val
			}
			if attr.Key == "content" {
				content = attr.Val
			}
		}
		if name != "" && content != "" {
			return content
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if keywords := extractKeywords(c); keywords != "" {
			return keywords
		}
	}
	return ""
}

// extractAuthor extracts author from meta tags
func extractAuthor(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var name, content string
		for _, attr := range n.Attr {
			if attr.Key == "name" && (attr.Val == "author" || attr.Val == "Author") {
				name = attr.Val
			}
			if attr.Key == "content" {
				content = attr.Val
			}
		}
		if name != "" && content != "" {
			return content
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if author := extractAuthor(c); author != "" {
			return author
		}
	}
	return ""
}

// extractOpenGraphTag extracts Open Graph meta tags
func extractOpenGraphTag(n *html.Node, property string) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var prop, content string
		for _, attr := range n.Attr {
			if attr.Key == "property" && attr.Val == property {
				prop = attr.Val
			}
			if attr.Key == "content" {
				content = attr.Val
			}
		}
		if prop == property && content != "" {
			return content
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if value := extractOpenGraphTag(c, property); value != "" {
			return value
		}
	}
	return ""
}

// extractImages extracts image URLs from HTML
func extractImages(n *html.Node, baseURL string) []string {
	var images []string
	var extract func(*html.Node)
	
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					if absoluteURL, err := toAbsoluteURL(attr.Val, baseURL); err == nil {
						images = append(images, absoluteURL)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	
	extract(n)
	return images
}

// extractVideos extracts video URLs from HTML
func extractVideos(n *html.Node, baseURL string) []string {
	var videos []string
	var extract func(*html.Node)
	
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "video" {
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						if absoluteURL, err := toAbsoluteURL(attr.Val, baseURL); err == nil {
							videos = append(videos, absoluteURL)
						}
					}
				}
			} else if n.Data == "source" {
				// Check if parent is video
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						srcVal := attr.Val
						if absoluteURL, err := toAbsoluteURL(srcVal, baseURL); err == nil {
							videos = append(videos, absoluteURL)
						}
					}
				}
			} else if n.Data == "iframe" {
				// Check for common video embeds (YouTube, Vimeo, etc.)
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						srcVal := attr.Val
						if isVideoEmbed(srcVal) {
							if absoluteURL, err := toAbsoluteURL(srcVal, baseURL); err == nil {
								videos = append(videos, absoluteURL)
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	
	extract(n)
	return videos
}

// isVideoEmbed checks if a URL is a known video embed
func isVideoEmbed(urlStr string) bool {
	lowerURL := strings.ToLower(urlStr)
	videoHosts := []string{"youtube.com", "youtu.be", "vimeo.com", "dailymotion.com", "twitch.tv"}
	for _, host := range videoHosts {
		if strings.Contains(lowerURL, host) {
			return true
		}
	}
	return false
}

// cleanHTML removes script and style tags from HTML
func cleanHTML(n *html.Node) {
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}
	
	var next *html.Node
	for c := n.FirstChild; c != nil; c = next {
		next = c.NextSibling
		cleanHTML(c)
	}
}
