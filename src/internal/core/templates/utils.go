package templates

import (
	"fmt"
	"strings"
)

// CSSVersion is appended to static CSS URLs for cache busting.
// Set this at startup (e.g. from the file's mod time).
var CSSVersion string

func CreatePageTitle(title string) string {
	if title == "" {
		title = "Home"
	}

	return fmt.Sprintf("%s | wax", title)
}

func FormatCallableID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}
