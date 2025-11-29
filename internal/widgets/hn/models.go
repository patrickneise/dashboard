package hn

// Item models the Hacker News "item" object.
// Docs (community): https://github.com/HackerNews/API
type Item struct {
	ID   int64  `json:"id"`
	Type string `json:"type"` // "story", "job", "comment", etc.

	By   string `json:"by"`
	Time int64  `json:"time"` // Unix time (seconds)

	Title string `json:"title"`
	URL   string `json:"url"`   // may be empty for Ask HN
	Text  string `json:"text"`  // may exist for Ask HN / Show HN
	Score int    `json:"score"` // points

	Descendents int64 `json:"descendants"` // comment count (stories)

	// Optional fields we might use later
	Kids []int64 `json:"kids"` // comment IDs
}
