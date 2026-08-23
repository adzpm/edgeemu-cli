package client

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/adzpm/edgeemu-cli/ds"
)

var itemRe = regexp.MustCompile(`(?s)<details data-name="[^"]*">\s*` +
	`<summary>(.*?)</summary>\s*` +
	`<p><a href="([^"]+)">download</a> \(<span>([^,]+), (\d+) DLs</span>\)</p>\s*` +
	`<p>system: <span>([^<]+)</span></p>\s*` +
	`<p>unpacked size: <span>([^<]+)</span></p>\s*` +
	`<p>hash: <span>([^<]*)</span></p>`)

// Search queries edgeemu.net for ROMs matching query within the given
// system ("all" searches every system).
func (c *Client) Search(ctx context.Context, query, system string) ([]ds.ROM, error) {
	form := url.Values{"search": {query}, "system": {system}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/search.php", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, err
	}

	var roms []ds.ROM

	for _, m := range itemRe.FindAllStringSubmatch(string(body), -1) {
		dls, _ := strconv.Atoi(m[4])
		roms = append(roms, ds.ROM{
			Name:         html.UnescapeString(m[1]),
			URL:          c.baseURL + html.UnescapeString(m[2]),
			Size:         strings.TrimSpace(m[3]),
			Downloads:    dls,
			System:       html.UnescapeString(m[5]),
			UnpackedSize: strings.TrimSpace(m[6]),
			Hash:         strings.TrimSpace(m[7]),
		})
	}

	return roms, nil
}
