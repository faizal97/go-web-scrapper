package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"time"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[SCRAPER] ", log.LstdFlags),
	}
}

type EnhancedScraper struct {
	client *http.Client
	logger *Logger
}

func NewEnhancedScraper() *EnhancedScraper {
	return &EnhancedScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: NewLogger(),
	}
}

func (es *EnhancedScraper) ScrapeWithContext(ctx context.Context, url string) ([]Story, error) {
	es.logger.Printf("starting scrape of %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Go-WebScraper-Tutorial/1.0")

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	stories := es.parseStories(doc)
	es.logger.Printf("finished scraping %s", url)

	return stories, nil
}

func (es *EnhancedScraper) parseStories(storyRow *goquery.Selection) Story {
	storyID := storyRow.Find("id")
	titleElement := storyRow.Find("span.titleline a").First()
	title := titleElement.Text()
	href, _ := titleElement.Attr("href")

	story := Story{
		Title: title,
		ID:    storyID,
	}

	if href != "" {
		if href[0] == '/' {
			story.URL = "https://news.ycombinator.com/" + href
		} else {
			story.URL = href
		}
	}

	metaRow := storyRow.Next()
	if !metaRow.HasClass("athing") {
		es.parseMetaData(&story, metaRow)
	}

	return story
}

func (es *EnhancedScraper) parseMetaData(story *Story, metaRow *goquery.Selection) {
	subtext := metaRow.Find("td.subtext")

	pointsText := subtext.Find("span.score").Text()
	if pointsText != "" {
		fmt.Sscanf(pointsText, "%d", &story.Points)
	}
	story.Author = subtext.Find("a.hnuser").Text()

	commentsLink := subtext.Find("a").Last()
	commentsText := commentsLink.Text()
	if commentsText != "" && commentsText != story.Author {
		fmt.Sscanf(commentsText, "%d", &story.Comments)
	}
}
