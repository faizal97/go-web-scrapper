# Go Web Scraper

A concurrent web scraper for Hacker News written in Go that allows you to scrape multiple pages simultaneously with configurable output formats.

## Tutorial

📖 **[Read the complete tutorial on Medium](https://faizalardianputra.medium.com/go-web-scraper-tutorial-building-a-concurrent-news-headlines-scraper-09d6ff376bf9)** - A step-by-step guide to building this concurrent web scraper from scratch.

## Features

- **Concurrent Scraping**: Scrape multiple pages simultaneously with configurable worker count
- **Multiple Output Formats**: Console display or JSON output
- **Graceful Shutdown**: Handles interrupts gracefully
- **Configurable Timeout**: Set request timeout limits
- **Verbose Logging**: Optional detailed logging
- **Story Metadata**: Extracts title, URL, points, comments, and author information

## Installation

```bash
git clone https://github.com/faizal97/go-web-scrapper.git
cd go-web-scrapper
go mod tidy
```

## Usage

### Basic Usage

```bash
# Build the scraper
go build -o scraper

# Scrape single page (default)
./scraper

# Scrape multiple pages with concurrent workers
./scraper -pages 3 -workers 5

# Output as JSON
./scraper -output json

# Enable verbose logging
./scraper -verbose
```

### Command Line Options

| Flag | Description | Default | Valid Range |
|------|-------------|---------|-------------|
| `-pages` | Number of pages to scrape | 1 | 1-10 |
| `-workers` | Number of concurrent workers | 3 | 1-10 |
| `-output` | Output format | console | console, json |
| `-timeout` | Request timeout in seconds | 30 | - |
| `-verbose` | Enable verbose logging | false | - |

### Examples

```bash
# Scrape 5 pages with 3 workers, output as JSON
./scraper -pages 5 -workers 3 -output json

# Scrape with verbose logging
./scraper -pages 2 -verbose

# Quick single page scrape
./scraper -pages 1
```

## Output Format

### Console Output
```
Found 30 stories:

1. Example Story Title
   Author: username | Points: 123 | Comments: 45
   URL: https://example.com

2. Another Story Title
   Author: user2 | Points: 89 | Comments: 12
   URL: https://news.ycombinator.com/item?id=123456
```

### JSON Output
```json
[
  {
    "Title": "Example Story Title",
    "URL": "https://example.com",
    "Points": 123,
    "Comments": 45,
    "Author": "username",
    "ID": "123456"
  }
]
```

## Dependencies

- `github.com/PuerkitoBio/goquery` - HTML parsing and DOM manipulation
- `github.com/andybalholm/cascadia` - CSS selector engine
- `golang.org/x/net` - Extended networking support

## Architecture

The scraper consists of several key components:

- **Scraper**: Basic HTTP client with timeout configuration
- **Story**: Data structure representing a Hacker News story
- **Config**: Configuration management with validation
- **Concurrent Scraper**: Handles multiple page scraping with worker pools
- **Enhanced Scraper**: Extended scraper with context support and logging

## License

This project is open source and available under the MIT License.