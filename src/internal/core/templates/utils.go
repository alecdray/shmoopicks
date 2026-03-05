package templates

import (
	"fmt"
	"strings"
)

func CreatePageTitle(title string) string {
	if title == "" {
		title = "Home"
	}

	return fmt.Sprintf("%s | wax", title)
}

func FormatCallableID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}
