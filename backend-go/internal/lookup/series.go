package lookup

import (
	"regexp"
	"strconv"
	"strings"
)

// seriesPattern matches a title that ends with a series-volume marker:
//
//	"<series> [, : ] (Vol|Volume|Vol.|Book|#) <number> [optional subtitle]"
//
// Group 1 is the series name; group 2 is the volume number. The optional
// trailing subtitle (introduced by ": <text>") is dropped.
//
// We deliberately require an explicit keyword (Vol/Volume/Book/#) — naked
// trailing numbers produce too many false positives ("Star Wars 1",
// "iOS 17", chapter numbers in fiction).
var seriesPattern = regexp.MustCompile(
	`(?i)^\s*(.+?)[\s,:]+(?:vol\.?|volume|book|#)\s*(\d+)\s*(?::.*)?\s*$`,
)

// ParseSeries returns the series name and volume number extracted from
// title, or ("", 0) if no series marker is present. Whitespace around
// the series name is collapsed and trimmed.
func ParseSeries(title string) (string, int) {
	m := seriesPattern.FindStringSubmatch(title)
	if m == nil {
		return "", 0
	}
	series := strings.TrimSpace(m[1])
	// Collapse internal runs of whitespace ("One   Piece" → "One Piece").
	series = strings.Join(strings.Fields(series), " ")
	if series == "" {
		return "", 0
	}
	vol, err := strconv.Atoi(m[2])
	if err != nil || vol <= 0 {
		return "", 0
	}
	return series, vol
}
