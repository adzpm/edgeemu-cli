package ds

import (
	"encoding/json"
	"testing"
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
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal into map: %v", err)
	}

	for _, key := range []string{"name", "system", "url", "size", "unpacked_size", "downloads", "hash"} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON output missing key %q (tag changed?): %s", key, data)
		}
	}

	var back ROM
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	if back != rom {
		t.Errorf("round trip mismatch:\ngot  %+v\nwant %+v", back, rom)
	}
}

func TestSystemJSONTags(t *testing.T) {
	sys := System{ID: "sega-genesis", Name: "Sega Mega Drive / Genesis"}

	data, err := json.Marshal(sys)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	want := `{"id":"sega-genesis","name":"Sega Mega Drive / Genesis"}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}
