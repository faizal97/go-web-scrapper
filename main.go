package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
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

	stories, err := scraper.ScrapeHackerNewsDetailed()
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

func (s *Scraper) ScrapeHackerNewsDetailed() ([]Story, error) {
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
			fmt.Sscanf(pointsText, "%d", &story.Points)
		}

		// Get Author
		story.Author = subText.Find("a.hnuser").Text()
		// Get Comments Count
		commentsLink := subText.Find("a").Last()
		commentsText := commentsLink.Text()
		if commentsText != "" && commentsText != story.Author {
			fmt.Sscanf(commentsText, "%d", &story.Comments)
		}

		stories = append(stories, story)
	})

	return stories, nil
}
