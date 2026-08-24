package ds

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestROMJSONTags(t *testing.T) {
	rom := ROM{
		Name:         "Sonic The Hedgehog (USA, Europe)",
		System:       "Sega Mega Drive / Genesis",
		URL:          "https://edgeemu.net/download/sega-genesis/x.zip",
		Size:         "377.87k",
		UnpackedSize: "512.00k",
		Downloads:    588,
		Hash:         "F9394E97",
	}

	data, err := json.Marshal(rom)
	require.NoError(t, err)

	var m map[string]any

	require.NoError(t, json.Unmarshal(data, &m))

	for _, key := range []string{"name", "system", "url", "size", "unpacked_size", "downloads", "hash"} {
		assert.Contains(t, m, key, "JSON tag changed?")
	}

	var back ROM

	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, rom, back, "JSON round trip must be lossless")
}

func TestSystemJSONTags(t *testing.T) {
	data, err := json.Marshal(System{ID: "sega-genesis", Name: "Sega Mega Drive / Genesis"})
	require.NoError(t, err)

	assert.JSONEq(t, `{"id":"sega-genesis","name":"Sega Mega Drive / Genesis"}`, string(data))
}
