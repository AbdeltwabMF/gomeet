package platform

import (
	"fmt"
	"net/url"
	"strings"
)

const ToolName = "gomeet"

// ValidateURL checks whether rawURL is a safe web URL for notifications and browsers.
func ValidateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) != rawURL {
		return fmt.Errorf("meeting URL contains surrounding whitespace")
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("parse meeting URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("meeting URL scheme %q is not allowed", u.Scheme)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("meeting URL must include a host")
	}
	if u.User != nil {
		return fmt.Errorf("meeting URL must not include user information")
	}

	return nil
}
