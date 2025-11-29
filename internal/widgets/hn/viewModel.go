package hn

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Rank     int
	Title    string
	URL      string
	Domain   string
	Score    int
	By       string
	Age      string
	Comments int
}

type WidgetViewModel struct {
	UpdatedAt string
	Entries   []Entry

	// stale indicator (set by widget.MarkStale)
	IsStale bool
	StaleBy string
}

// BuildViewModel converts items (in rank order) into a render-friendly view model.
func BuildViewModel(now time.Time, items []*Item) WidgetViewModel {
	entries := make([]Entry, 0, len(items))

	for i, it := range items {
		if it == nil {
			continue
		}

		link := itemURL(it)
		entries = append(entries, Entry{
			Rank:     i + 1,
			Title:    it.Title,
			URL:      link,
			Domain:   domainFromURL(it.URL),
			Score:    it.Score,
			By:       it.By,
			Age:      relativeAge(now, it.Time),
			Comments: int(it.Descendents),
		})
	}

	return WidgetViewModel{
		UpdatedAt: now.Format("3:04 PM"),
		Entries:   entries,
	}
}

// itemURL links to the external URL when present; otherwise links to the HN item page (Ask HN, etc.).
func itemURL(it *Item) string {
	if it.URL != "" {
		return it.URL
	}
	return "https://news.ycombinator.com/item?id=" + strconv.FormatInt(it.ID, 10)
}

func domainFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	return strings.TrimPrefix(u.Host, "www.")
}

func relativeAge(now time.Time, unixSeconds int64) string {
	t := time.Unix(unixSeconds, 0)
	if t.After(now) {
		return "just now"
	}

	d := now.Sub(t)
	switch {
	case d < 1*time.Minute:
		return "just now"
	case d < 1*time.Hour:
		return short(int(d.Minutes()), "m")
	case d < 24*time.Hour:
		return short(int(d.Hours()), "h")
	default:
		return short(int(d.Hours()/24), "d")
	}
}

func short(n int, suffix string) string {
	if n <= 1 {
		return "1" + suffix
	}
	return strconv.Itoa(n) + suffix
}
