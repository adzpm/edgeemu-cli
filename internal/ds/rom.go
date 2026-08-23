package ds

// ROM is a single search result from edgeemu.net.
type ROM struct {
	Name         string `json:"name"`
	System       string `json:"system"`
	URL          string `json:"url"`
	Size         string `json:"size"`
	UnpackedSize string `json:"unpacked_size"`
	Downloads    int    `json:"downloads"`
	Hash         string `json:"hash"`
}
