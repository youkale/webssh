package tui

import (
	"fmt"
	"mime"
	"strconv"
	"strings"
	"time"
)

// orStr returns string b if string a is empty, otherwise returns a
func orStr(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

const (
	_  = 1 << (10 * iota) // ignore first value by assigning to blank identifier
	KB                    // 1 << (10*1)
	MB                    // 1 << (10*2)
	GB                    // 1 << (10*3)
	TB                    // 1 << (10*4)
)

func humanBytes(i int64) string {
	bytes := float64(i)
	switch {
	case bytes < KB:
		return fmt.Sprintf("%.0fB", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.1fK", bytes/KB)
	case bytes < GB:
		return fmt.Sprintf("%.1fM", bytes/MB)
	case bytes < TB:
		return fmt.Sprintf("%.1fG", bytes/GB)
	default:
		return fmt.Sprintf("%.1fT", bytes/TB)
	}
}

// humanSize converts a string containing bytes size to human readable format
// Returns empty string if input is invalid
func humanSize(size string) string {
	atoi, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return ""
	}
	return humanBytes(atoi)
}

// HumanMillis converts milliseconds to a human-readable duration string
func humanMillis(ms int64) string {
	duration := time.Duration(ms) * time.Millisecond

	if duration < time.Second {
		return fmt.Sprintf("%dms", ms)
	}

	if duration < time.Minute {
		return fmt.Sprintf("%.1fs", float64(duration)/float64(time.Second))
	}

	if duration < time.Hour {
		minutes := duration / time.Minute
		seconds := (duration % time.Minute) / time.Second
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// parseContentType extracts and formats the media type from a Content-Type header
func parseContentType(contentType string) string {
	if contentType == "" {
		return "-"
	}
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "-"
	}

	// Simplify content type for display
	switch {
	case strings.Contains(mediatype, "application/json"):
		return "json"
	case strings.Contains(mediatype, "text/html"):
		return "html"
	case strings.Contains(mediatype, "text/plain"):
		return "text"
	case strings.Contains(mediatype, "application/xml"):
		return "xml"
	case strings.Contains(mediatype, "form"):
		return "form"
	default:
		if len(mediatype) > 15 {
			return mediatype[:12] + "..."
		}
		return mediatype
	}
}
