package main

import (
	"fmt"
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
