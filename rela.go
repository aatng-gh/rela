package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-shiori/go-readability"
)

type Article struct {
	Title       string // title of article
	RawContent  string // raw content, e.g., with HTML or formatting
	TextContent string // plain text content, stripped of any formatting
	Length      int    // length of plain text content
}

type Parser interface {
	Parse(io.Reader, *url.URL) (Article, error)
}

type ReadabilityParser struct{}

func (ReadabilityParser) Parse(body io.Reader, parsedURL *url.URL) (Article, error) {
	a, err := readability.FromReader(body, parsedURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse document from %s: %w", parsedURL.String(), err)
	}
	return Article{Title: a.Title, RawContent: a.Content, TextContent: a.TextContent, Length: a.Length}, nil
}

func ParseArticleFromURL(rawURL string, client *http.Client, parser ReadabilityParser, logger *slog.Logger) (Article, error) {
	logger.Info("parsing article from url...")

	resp, err := http.Get(rawURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse url: %w", err)
	}

	article, err := parser.Parse(resp.Body, parsedURL)
	if err != nil {
		return Article{}, fmt.Errorf("failed to parse response body: %w", err)
	}

	logger.Info("successfully parsed article")
	return article, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	parser := ReadabilityParser{}

	urls := []string{
		"https://cbea.ms/git-commit/",
		"https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html",
	}

	for _, u := range urls {
		childLogger := logger.With("url", u)

		_, err := ParseArticleFromURL(u, client, parser, childLogger)
		if err != nil {
			childLogger.Error("failed to parse article from url", slog.String("error", err.Error()))
		}
	}
}
