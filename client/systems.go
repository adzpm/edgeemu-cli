package client

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"

	"github.com/adzpm/edgeemu-cli/ds"
)

var optionRe = regexp.MustCompile(`<option value="([^"]+)"[^>]*>([^<]+)</option>`)

// Systems fetches the list of searchable systems from edgeemu.net.
func (c *Client) Systems(ctx context.Context) ([]ds.System, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/search.php", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
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

	return systems, nil
}
