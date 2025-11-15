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
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractTitle(t *testing.T) {
	htmlContent := `<html><head><title>Test Page</title></head><body></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	title := extractTitle(doc)
	if title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", title)
	}
}

func TestExtractMetaDescription(t *testing.T) {
	htmlContent := `<html><head><meta name="description" content="This is a test description"></head><body></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	desc := extractMetaDescription(doc)
	if desc != "This is a test description" {
		t.Errorf("Expected description 'This is a test description', got '%s'", desc)
	}
}

func TestExtractKeywords(t *testing.T) {
	htmlContent := `<html><head><meta name="keywords" content="golang, search, engine"></head><body></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	keywords := extractKeywords(doc)
	if keywords != "golang, search, engine" {
		t.Errorf("Expected keywords 'golang, search, engine', got '%s'", keywords)
	}
}

func TestExtractAuthor(t *testing.T) {
	htmlContent := `<html><head><meta name="author" content="John Doe"></head><body></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	author := extractAuthor(doc)
	if author != "John Doe" {
		t.Errorf("Expected author 'John Doe', got '%s'", author)
	}
}

func TestExtractOpenGraphTag(t *testing.T) {
	htmlContent := `<html><head><meta property="og:title" content="OG Title"></head><body></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	ogTitle := extractOpenGraphTag(doc, "og:title")
	if ogTitle != "OG Title" {
		t.Errorf("Expected OG title 'OG Title', got '%s'", ogTitle)
	}
}

func TestExtractImages(t *testing.T) {
	htmlContent := `<html><body><img src="image1.jpg"><img src="image2.png"></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	images := extractImages(doc, "https://example.com")
	if len(images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(images))
	}
}

func TestExtractVideos(t *testing.T) {
	htmlContent := `<html><body>
		<video src="video1.mp4"></video>
		<iframe src="https://www.youtube.com/embed/abc123"></iframe>
	</body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	videos := extractVideos(doc, "https://example.com")
	if len(videos) < 1 {
		t.Errorf("Expected at least 1 video, got %d", len(videos))
	}
}

func TestIsVideoEmbed(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.youtube.com/embed/abc123", true},
		{"https://youtu.be/abc123", true},
		{"https://vimeo.com/123456", true},
		{"https://example.com/video.mp4", false},
		{"https://example.com", false},
	}
	
	for _, test := range tests {
		result := isVideoEmbed(test.url)
		if result != test.expected {
			t.Errorf("For URL '%s', expected %v, got %v", test.url, test.expected, result)
		}
	}
}

func TestCleanHTML(t *testing.T) {
	htmlContent := `<html><head><script>alert('test');</script></head><body><style>body{}</style><p>Content</p></body></html>`
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	cleanHTML(doc)
	
	// Check that script and style tags are removed
	var hasScript, hasStyle bool
	var checkNodes func(*html.Node)
	checkNodes = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" {
				hasScript = true
			}
			if n.Data == "style" {
				hasStyle = true
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			checkNodes(c)
		}
	}
	checkNodes(doc)
	
	if hasScript {
		t.Error("Expected script tags to be removed")
	}
	if hasStyle {
		t.Error("Expected style tags to be removed")
	}
}

func TestExtractEnhancedMetadata(t *testing.T) {
	htmlContent := `<html>
		<head>
			<title>Test Page</title>
			<meta name="description" content="Test description">
			<meta name="keywords" content="test, keywords">
			<meta name="author" content="Test Author">
			<meta property="og:title" content="OG Test Title">
			<meta property="og:description" content="OG Test Description">
			<meta property="og:image" content="https://example.com/og-image.jpg">
		</head>
		<body>
			<img src="image1.jpg">
			<img src="image2.jpg">
		</body>
	</html>`
	
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	
	metadata := extractEnhancedMetadata(doc, "https://example.com")
	
	if metadata["title"] != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%v'", metadata["title"])
	}
	
	if metadata["meta_description"] != "Test description" {
		t.Errorf("Expected description 'Test description', got '%v'", metadata["meta_description"])
	}
	
	if metadata["keywords"] != "test, keywords" {
		t.Errorf("Expected keywords 'test, keywords', got '%v'", metadata["keywords"])
	}
	
	if metadata["author"] != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%v'", metadata["author"])
	}
	
	if metadata["og_title"] != "OG Test Title" {
		t.Errorf("Expected og_title 'OG Test Title', got '%v'", metadata["og_title"])
	}
	
	if metadata["og_description"] != "OG Test Description" {
		t.Errorf("Expected og_description 'OG Test Description', got '%v'", metadata["og_description"])
	}
	
	if metadata["og_image"] != "https://example.com/og-image.jpg" {
		t.Errorf("Expected og_image 'https://example.com/og-image.jpg', got '%v'", metadata["og_image"])
	}
	
	// Check that image_urls is a JSON array
	if imageURLs, ok := metadata["image_urls"].(string); ok {
		if imageURLs == "" || imageURLs == "[]" {
			t.Error("Expected image_urls to contain images")
		}
	} else {
		t.Error("Expected image_urls to be a string")
	}
}
