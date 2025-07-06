package main

import (
	"flag"
	"fmt"
)

type Config struct {
	Pages   int
	Workers int
	Output  string
	Timeout int
	Verbose bool
}

func ParseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Pages, "pages", 1, "Number of pages to scrape")
	flag.IntVar(&config.Workers, "workers", 3, "Number of concurrent workers")
	flag.StringVar(&config.Output, "output", "console", "Output Format(console, json)")
	flag.IntVar(&config.Timeout, "timeout", 30, "request timeout in seconds")
	flag.BoolVar(&config.Verbose, "verbose", false, "enable verbose logging")
	flag.Parse()

	return config
}

func (c *Config) Validate() error {
	if c.Pages < 1 || c.Pages > 10 {
		return fmt.Errorf("pages must be between 1 and 10")
	}

	if c.Workers < 1 || c.Workers > 10 {
		return fmt.Errorf("workers must be between 1 and 10")
	}

	if c.Output != "console" && c.Output != "json" {
		return fmt.Errorf("output must be 'console' or 'json'")
	}

	return nil
}
