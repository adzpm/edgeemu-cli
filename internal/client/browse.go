package client

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// ErrNoLetters is returned when a browse page carries no letter buckets,
// which most likely means the page layout has changed.
var ErrNoLetters = errors.New("no letter buckets parsed: page layout may have changed")

// browseItemRe matches one ROM entry on a /browse page. Unlike search
// results there is no "system:" line (the page is already per-system),
// and the hash label uses &nbsp; instead of a space.
var browseItemRe = regexp.MustCompile(`(?s)<details data-name="[^"]*">\s*` +
	`<summary>(.*?)</summary>\s*` +
	`<p><a href="([^"]+)">download</a> \(<span>([^,]+), (\d+) DLs</span>\)</p>\s*` +
	`<p>unpacked size: <span>([^<]+)</span></p>\s*` +
	`<p>hash:(?:&nbsp;|\s)*<span>([^<]*)</span></p>`)

// BrowseLetters fetches the letter buckets available for a system on
// its /browse page (typically "1" and "a" through "z").
func (c *Client) BrowseLetters(ctx context.Context, system string) ([]string, error) {
	body, err := c.get(ctx, "/browse/"+system)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`value="/browse/` + regexp.QuoteMeta(system) + `/([^"/]+)"`)

	var letters []string
	for _, m := range re.FindAllStringSubmatch(string(body), -1) {
		letters = append(letters, m[1])
	}

	if len(letters) == 0 {
		return nil, fmt.Errorf("%w (%s/browse/%s)", ErrNoLetters, c.baseURL, system)
	}

	return letters, nil
}

// Browse fetches every ROM of one letter bucket of a system. Browse
// pages are not capped at 100 entries like the search is. The System
// field is left empty: the browse page does not repeat it.
func (c *Client) Browse(ctx context.Context, system, letter string) ([]ds.ROM, error) {
	body, err := c.get(ctx, "/browse/"+system+"/"+letter)
	if err != nil {
		return nil, err
	}

	var roms []ds.ROM

	for _, m := range browseItemRe.FindAllStringSubmatch(string(body), -1) {
		dls, _ := strconv.Atoi(m[4])
		roms = append(roms, ds.ROM{
			Name:         html.UnescapeString(m[1]),
			URL:          c.baseURL + html.UnescapeString(m[2]),
			Size:         strings.TrimSpace(m[3]),
			Downloads:    dls,
			System:       "",
			UnpackedSize: strings.TrimSpace(m[5]),
			Hash:         strings.TrimSpace(m[6]),
		})
	}

	return roms, nil
}

// get performs a GET request against the site and returns the body.
func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: %w: %s", path, ErrRequestFailed, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, err
	}

	return body, nil
}
