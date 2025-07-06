package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Story represents a news story
type Story struct {
	Title    string
	URL      string
	Points   int
	Comments int
	Author   string
	ID       string
}

// Scraper holds our scrapping configuration
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
	config := ParseFlags()

	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down...")
		cancel()
	}()

	if err := runScraper(ctx, config); err != nil {
		log.Fatal(err)
	}
}

func runScraper(ctx context.Context, config *Config) error {
	scraper := NewEnhancedScraper()

	if config.Verbose {
		scraper.logger.Printf("Starting scraper with config:pages=%d, workers=%d", config.Pages, config.Workers)
	}

	start := time.Now()
	var allStories []Story

	if config.Pages == 1 {
		stories, err := scraper.ScrapeWithContext(ctx, "https://news.ycombinator.com")
		if err != nil {
			return fmt.Errorf("error scraping: %w", err)
		}
		allStories = stories
	}

	duration := time.Since(start)

	if config.Verbose {
		scraper.logger.Printf("scraping completed in %v, found %d stories", duration, len(allStories))
	}

	return outputStories(allStories, config)
}

func outputStories(stories []Story, config *Config) error {
	switch config.Output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(stories)
	case "console":
		fallthrough
	default:
		displayStories(stories)
		return nil
	}
}

func basicScraping() {
	scraper := NewScrapper()

	stories, err := scraper.ScrapeHackerNewsDetailed()
	if err != nil {
		log.Fatal("error scraping:", err)
	}

	displayStories(stories)
}

func concurrentScraping() {
	scraper := NewConcurrentScraper(3)
	fmt.Println("starting concurrent scraping")
	start := time.Now()

	stories, err := scraper.ScrapeMultiplePages(3)
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("\nScraping completed in %v\n", duration)

	displayStories(stories)
}

func displayStories(stories []Story) {
	fmt.Printf("\nFound %d stories:\n\n", len(stories))
	for i, story := range stories {
		if i >= 20 {
			fmt.Printf("...and %d more stories\n", len(stories)-i)
			break
		}

		fmt.Printf("%d. %s\n", i+1, story.Title)
		if story.Author != "" || story.Points > 0 || story.Comments > 0 {
			fmt.Printf("Author: %s | Points: %d | Comments: %d\n", story.Author, story.Points, story.Comments)
		}
		if story.URL != "" {
			fmt.Printf("URL: %s\n", story.URL)
		}
		fmt.Println()
	}
}

func (s *Scraper) ScrapeHackerNewsDetailed() ([]Story, error) {
	url := "https://news.ycombinator.com"

	//	Make HTTP Request
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	// Check Status Code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var stories []Story
	// Find all story elements
	doc.Find("tr.athing").Each(func(i int, storyRow *goquery.Selection) {
		// Get Story ID
		storyID, _ := storyRow.Attr("id")

		// Get Title and URL
		titleElement := storyRow.Find("span.titleline a").First()
		title := titleElement.Text()
		href, _ := titleElement.Attr("href")

		if title == "" {
			return
		}

		story := Story{
			Title: title,
			ID:    storyID,
		}

		// Handle URLs
		if href != "" {
			if strings.HasPrefix(href, "/") {
				story.URL = "https://news.ycombinator.com" + href
			} else {
				story.URL = href
			}
		}

		// Get MetaData from the next row(subtext)
		metaRow := storyRow.Next()
		if metaRow.HasClass("athing") {
			// This Story doesn't have metadata (like job posts)
			stories = append(stories, story)
		}

		subText := metaRow.Find("td.subtext")

		// Get Points
		pointsText := subText.Find("span.score").Text()
		if pointsText == "" {
			_, err := fmt.Sscanf(pointsText, "%d", &story.Points)
			if err != nil {
				return
			}
		}

		// Get Author
		story.Author = subText.Find("a.hnuser").Text()
		// Get Comments Count
		commentsLink := subText.Find("a").Last()
		commentsText := commentsLink.Text()
		if commentsText != "" && commentsText != story.Author {
			_, err := fmt.Sscanf(commentsText, "%d", &story.Comments)
			if err != nil {
				return
			}
		}

		stories = append(stories, story)
	})

	return stories, nil
}
