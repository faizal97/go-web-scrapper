package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"time"
)

// Story represents a news story
type Story struct {
	Title string
	URL   string
	ID    int
}

// Scrapper holds our scrapping configuration
type Scraper struct {
	client *http.Client
}

func NewScrapper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func main() {
	scraper := NewScrapper()

	stories, err := scraper.ScrapeHackerNews()
	if err != nil {
		log.Fatal("Error scraping:", err)
	}

	fmt.Printf("Found %d stories:\n\n", len(stories))
	for i, story := range stories {
		fmt.Printf("%d. %s\n", i+1, story.Title)
		if story.URL != "" {
			fmt.Printf("URL:%s\n", story.URL)
		}
		fmt.Println()
	}
}

func (s *Scraper) ScrapeHackerNews() ([]Story, error) {
	url := "https://news.ycombinator.com"

	//	Make HTTP Request
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	// Check Status Code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad status code: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse HTML: %w", err)
	}

	var stories []Story
	// Find all story elements
	doc.Find("tr.athing").Each(func(i int, s *goquery.Selection) {
		titleElement := s.Find("span.titleline a").First()
		title := titleElement.Text()
		href, exists := titleElement.Attr("href")

		if title != "" {
			story := Story{
				Title: title,
			}

			// Handle relative URLs
			if exists {
				if href[0] == '/' {
					story.URL = "https://news.ycombinator.com" + href
				} else {
					story.URL = href
				}
			}
			stories = append(stories, story)
		}
	})

	return stories, nil
}
