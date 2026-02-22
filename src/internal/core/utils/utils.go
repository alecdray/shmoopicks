package utils

import (
	"fmt"
)

func CreatePageTitle(title string) string {
	if title == "" {
		title = "Home"
	}

	return fmt.Sprintf("%s | shmoopicks", title)
}
