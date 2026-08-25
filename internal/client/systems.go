package client

import (
	"context"
	"fmt"
	"html"
	"regexp"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

var optionRe = regexp.MustCompile(`<option value="([^"]+)"[^>]*>([^<]+)</option>`)

// Systems fetches the list of searchable systems from edgeemu.net.
func (c *Client) Systems(ctx context.Context) ([]ds.System, error) {
	body, err := c.get(ctx, "/search.php")
	if err != nil {
		return nil, err
	}

	var systems []ds.System

	for _, m := range optionRe.FindAllStringSubmatch(string(body), -1) {
		if m[1] == "all" {
			continue
		}

		systems = append(systems, ds.System{ID: m[1], Name: html.UnescapeString(m[2])})
	}

	// Zero systems means the page layout changed and parsing broke;
	// erroring here also keeps the empty list out of the cache.
	if len(systems) == 0 {
		return nil, fmt.Errorf("%w (%s/search.php)", ErrNoSystems, c.baseURL)
	}

	return systems, nil
}
