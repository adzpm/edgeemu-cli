// Package ds holds the domain structures shared across the CLI.
package ds

// ROM is a single search result from edgeemu.net.
type ROM struct {
	Name         string `json:"name"          xml:"name"          yaml:"name"`
	System       string `json:"system"        xml:"system"        yaml:"system"`
	URL          string `json:"url"           xml:"url"           yaml:"url"`
	Size         string `json:"size"          xml:"size"          yaml:"size"`
	UnpackedSize string `json:"unpacked_size" xml:"unpacked_size" yaml:"unpacked_size"`
	Downloads    int    `json:"downloads"     xml:"downloads"     yaml:"downloads"`
	Hash         string `json:"hash"          xml:"hash"          yaml:"hash"`
}
