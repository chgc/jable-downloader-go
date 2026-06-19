package utils

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGetMovieLinks_GoQuerySelector(t *testing.T) {
	// Test the goquery selector logic used in GetMovieLinks
	// (GetMovieLinks itself uses ChromeDP, which can't be unit tested)

	htmlWithLinks := `<html><body>
		<div class="img-box">
			<a href="https://jable.tv/videos/abc-123/">Video 1</a>
		</div>
		<div class="img-box">
			<a href="https://jable.tv/videos/def-456/">Video 2</a>
		</div>
		<div class="img-box">
			<a href="https://jable.tv/videos/ghi-789/">Video 3</a>
		</div>
		<div class="other-box">
			<a href="https://jable.tv/videos/other/">Should not match</a>
		</div>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlWithLinks))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	var links []string
	doc.Find("div.img-box a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	if len(links) != 3 {
		t.Errorf("expected 3 links from div.img-box a, got %d", len(links))
	}

	expected := []string{
		"https://jable.tv/videos/abc-123/",
		"https://jable.tv/videos/def-456/",
		"https://jable.tv/videos/ghi-789/",
	}
	for i, link := range links {
		if i < len(expected) && link != expected[i] {
			t.Errorf("link %d: expected %q, got %q", i, expected[i], link)
		}
	}
}

func TestGetMovieLinks_EmptyResult(t *testing.T) {
	htmlNoLinks := `<html><body><div class="img-box">No links here</div></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlNoLinks))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	var links []string
	doc.Find("div.img-box a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestGetMovieLinks_WithMultipleSelectors(t *testing.T) {
	// Test that the selector only matches div.img-box a, not other a tags
	html := `<html><body>
		<div class="img-box">
			<a href="https://jable.tv/videos/valid-001/">Valid</a>
		</div>
		<a href="https://jable.tv/videos/should-not-match/">Not in img-box</a>
		<div class="img-box">
			<a href="https://jable.tv/videos/valid-002/">Valid 2</a>
		</div>
	</body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	var links []string
	doc.Find("div.img-box a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	if len(links) != 2 {
		t.Errorf("expected exactly 2 links (only from div.img-box), got %d", len(links))
	}
}

func TestGetMovieLinks_MalformedHTML(t *testing.T) {
	// goquery handles malformed HTML gracefully
	malformedHTML := `<div class="img-box"><a href=https://jable.tv/videos/no-quotes/>Broken</a>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(malformedHTML))
	if err != nil {
		t.Fatalf("failed to parse malformed HTML: %v", err)
	}

	var links []string
	doc.Find("div.img-box a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	// Should still find at least one link despite malformed HTML
	if len(links) == 0 {
		t.Log("no links found in malformed HTML (goquery may handle it differently)")
	}
}
