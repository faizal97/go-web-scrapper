package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"sync"
	"time"
)

// ConcurrentScraper handles concurrent scraping operations
type ConcurrentScraper struct {
	scraper *Scraper
	workers int
}

// NewConcurrentScraper creates a new concurrent scraper
func NewConcurrentScraper(workers int) *ConcurrentScraper {
	return &ConcurrentScraper{
		scraper: NewScraper(),
		workers: workers,
	}
}

// ScrapeResult holds the result of a scraping operation
type ScrapeResult struct {
	Stories []Story
	Page    int
	Error   error
}

// ScrapeMultiplePages scrapes multiple pages concurrently
func (cs *ConcurrentScraper) ScrapeMultiplePages(pages int) ([]Story, error) {
	jobs := make(chan int, pages)
	results := make(chan ScrapeResult, pages)

	var wg sync.WaitGroup
	for w := 0; w < cs.workers; w++ {
		wg.Add(1)
		go cs.worker(jobs, results, &wg)
	}

	for p := 1; p <= pages; p++ {
		jobs <- p
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allStories []Story
	var errors []error

	for result := range results {
		if result.Error != nil {
			errors = append(errors, result.Error)
		} else {
			allStories = append(allStories, result.Stories...)
		}
	}

	if len(errors) > 0 {
		return allStories, fmt.Errorf("encountered %d errors", len(errors))
	}

	return allStories, nil
}

func (cs *ConcurrentScraper) worker(jobs <-chan int, result chan<- ScrapeResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for page := range jobs {
		fmt.Printf("scraping page %d...\n", page)

		var url string
		if page == 1 {
			url = "https://news.ycombinator.com"
		} else {
			url = fmt.Sprintf("https://news.ycombinator.com/news?p=%d", page)
		}

		stories, err := cs.scrapePageByURL(url)

		result <- ScrapeResult{
			Stories: stories,
			Page:    page,
			Error:   err,
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (cs *ConcurrentScraper) scrapePageByURL(url string) ([]Story, error) {
	resp, err := cs.scraper.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stories: %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code for %s: %d", url, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from %s: %w", url, err)
	}

	var stories []Story
	doc.Find("tr.athing").Each(func(i int, storyRow *goquery.Selection) {
		storyID, _ := storyRow.Attr("id")
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

		if href != "" {
			if href[0] == '/' {
				story.URL = "https://news.ycombinator.com" + href
			} else {
				story.URL = href
			}
		}

		metaRow := storyRow.Next()
		if !metaRow.HasClass("athing") {
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

		stories = append(stories, story)
	})

	return stories, nil
}

// NewConcurrentScraperWithContext creates a new concurrent scraper with enhanced scraper
func NewConcurrentScraperWithContext(scraper *EnhancedScraper, workers int) *ConcurrentScraperWithContext {
	return &ConcurrentScraperWithContext{
		scraper: scraper,
		workers: workers,
	}
}

// ConcurrentScraperWithContext handles concurrent scraping with context
type ConcurrentScraperWithContext struct {
	scraper *EnhancedScraper
	workers int
}

// ScrapeMultiplePagesWithContext scrapes multiple pages concurrently with context
func (cs *ConcurrentScraperWithContext) ScrapeMultiplePagesWithContext(ctx context.Context, pages int) ([]Story, error) {
	jobs := make(chan int, pages)
	results := make(chan ScrapeResult, pages)

	var wg sync.WaitGroup
	for w := 0; w < cs.workers; w++ {
		wg.Add(1)
		go cs.workerWithContext(ctx, jobs, results, &wg)
	}

	for p := 1; p <= pages; p++ {
		jobs <- p
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allStories []Story
	var errors []error

	for result := range results {
		if result.Error != nil {
			errors = append(errors, result.Error)
		} else {
			allStories = append(allStories, result.Stories...)
		}
	}

	if len(errors) > 0 {
		return allStories, fmt.Errorf("encountered %d errors during scraping", len(errors))
	}

	return allStories, nil
}

// workerWithContext processes scraping jobs with context
func (cs *ConcurrentScraperWithContext) workerWithContext(ctx context.Context, jobs <-chan int, results chan<- ScrapeResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for page := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Printf("Scraping page %d...\n", page)

			var url string
			if page == 1 {
				url = "https://news.ycombinator.com"
			} else {
				url = fmt.Sprintf("https://news.ycombinator.com/news?p=%d", page)
			}

			stories, err := cs.scraper.ScrapeWithContext(ctx, url)

			results <- ScrapeResult{
				Stories: stories,
				Page:    page,
				Error:   err,
			}

			// Be nice to the server
			time.Sleep(100 * time.Millisecond)
		}
	}
}
