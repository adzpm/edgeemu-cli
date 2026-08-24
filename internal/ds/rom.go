package ds

// ROM is a single search result from edgeemu.net.
type ROM struct {
	Name         string `json:"name" yaml:"name" xml:"name"`
	System       string `json:"system" yaml:"system" xml:"system"`
	URL          string `json:"url" yaml:"url" xml:"url"`
	Size         string `json:"size" yaml:"size" xml:"size"`
	UnpackedSize string `json:"unpacked_size" yaml:"unpacked_size" xml:"unpacked_size"`
	Downloads    int    `json:"downloads" yaml:"downloads" xml:"downloads"`
	Hash         string `json:"hash" yaml:"hash" xml:"hash"`
}
