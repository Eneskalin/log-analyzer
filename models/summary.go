package models

type LogSummary struct {
	FileName      string
	TotalLines    int
	MatchedEvents int
	SeverityStats map[string]int
	Details       []string
}
