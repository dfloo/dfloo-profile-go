package scraper

import (
	"context"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type JobPostingSnapshot struct {
	URL       string            `json:"url"`
	FinalURL  string            `json:"finalUrl"`
	Title     string            `json:"title"`
	Meta      map[string]string `json:"meta,omitempty"`
	Text      string            `json:"text,omitempty"`
	HTML      string            `json:"html,omitempty"`
	FetchedAt time.Time         `json:"fetchedAt"`
}

type JobPostingExtract struct {
	Company     string
	Role        string
	Description string
}

func ScrapeJobPosting(ctx context.Context, targetURL string) (JobPostingExtract, JobPostingSnapshot, error) {
	snapshot := JobPostingSnapshot{
		URL:       targetURL,
		FetchedAt: time.Now().UTC(),
		Meta:      map[string]string{},
	}
	extract := JobPostingExtract{}

	collector := colly.NewCollector(
		colly.UserAgent("dfloo-profile-go/1.0"),
	)
	collector.SetRequestTimeout(12 * time.Second)

	var title string
	var h1Text string
	var bodyText string
	var htmlText string
	var lastErr error

	collector.OnRequest(func(r *colly.Request) {
		if ctx.Err() != nil {
			r.Abort()
			lastErr = ctx.Err()
		}
	})

	collector.OnResponse(func(r *colly.Response) {
		snapshot.FinalURL = r.Request.URL.String()
		htmlText = string(r.Body)
	})

	collector.OnHTML("title", func(e *colly.HTMLElement) {
		if title == "" {
			title = strings.TrimSpace(e.Text)
		}
	})

	collector.OnHTML("h1", func(e *colly.HTMLElement) {
		if h1Text == "" {
			h1Text = strings.TrimSpace(e.Text)
		}
	})

	collector.OnHTML("meta", func(e *colly.HTMLElement) {
		name := strings.ToLower(strings.TrimSpace(e.Attr("name")))
		prop := strings.ToLower(strings.TrimSpace(e.Attr("property")))
		content := strings.TrimSpace(e.Attr("content"))
		if content == "" {
			return
		}
		key := name
		if key == "" {
			key = prop
		}
		if key != "" {
			snapshot.Meta[key] = content
		}
	})

	collector.OnHTML("body", func(e *colly.HTMLElement) {
		if bodyText == "" {
			bodyText = strings.TrimSpace(e.DOM.Text())
		}
	})

	collector.OnError(func(_ *colly.Response, err error) {
		lastErr = err
	})

	if err := collector.Visit(targetURL); err != nil {
		return extract, snapshot, err
	}

	if lastErr != nil {
		return extract, snapshot, lastErr
	}

	snapshot.Title = title
	snapshot.Text = limitString(bodyText, 12000)
	snapshot.HTML = limitString(htmlText, 250000)

	extract.Description = firstNonEmpty(
		snapshot.Meta["og:description"],
		snapshot.Meta["description"],
		limitString(snapshot.Text, 4000),
	)

	roleTitle := firstNonEmpty(
		snapshot.Meta["og:title"],
		snapshot.Meta["twitter:title"],
		title,
		h1Text,
	)

	company := firstNonEmpty(
		snapshot.Meta["og:site_name"],
		snapshot.Meta["company"],
	)

	role, parsedCompany := splitRoleCompany(roleTitle)
	if company == "" {
		company = parsedCompany
	}

	extract.Role = role
	extract.Company = company

	return extract, snapshot, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func splitRoleCompany(title string) (string, string) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ""
	}

	separators := []string{" at ", " @ ", " - ", " | ", " – ", " — "}
	for _, sep := range separators {
		if strings.Contains(title, sep) {
			parts := strings.SplitN(title, sep, 2)
			role := strings.TrimSpace(parts[0])
			company := strings.TrimSpace(parts[1])
			return role, company
		}
	}

	return title, ""
}

func limitString(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
