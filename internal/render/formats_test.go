package render

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

var testROMs = []ds.ROM{
	{Name: "Sonic & Knuckles (World)", System: "Sega Mega Drive / Genesis", URL: "https://example.com/1.zip", Size: "1.36m", UnpackedSize: "256.00k", Downloads: 341, Hash: "4DCFD55C 0658F691"},
	{Name: "Sonic The Hedgehog 2 (World)", System: "Sega Mega Drive / Genesis", URL: "https://example.com/2.zip", Size: "732.08k", UnpackedSize: "1.00m", Downloads: 432, Hash: "24AB4C3A"},
}

var testSystems = []ds.System{
	{ID: "atari-2600", Name: "Atari 2600"},
	{ID: "sega-genesis", Name: "Sega Mega Drive / Genesis"},
}

func TestJSONROMs(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).JSONROMs(testROMs, []string{"name", "dls"}))

	var got []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))

	require.Len(t, got, 2)
	assert.Equal(t, map[string]any{"name": "Sonic & Knuckles (World)", "dls": float64(341)}, got[0])

	// Field order must be canonical, not alphabetical.
	assert.Less(t, strings.Index(buf.String(), `"name"`), strings.Index(buf.String(), `"dls"`))
}

func TestJSONROMsEmptyIsArray(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).JSONROMs(nil, []string{"name"}))

	assert.Equal(t, "[]", strings.TrimSpace(buf.String()))
}

func TestYAMLROMs(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).YAMLROMs(testROMs, []string{"name", "size", "dls"}))

	var got []map[string]any
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &got))

	require.Len(t, got, 2)
	assert.Equal(t, "Sonic & Knuckles (World)", got[0]["name"])
	assert.Equal(t, 341, got[0]["dls"], "dls must stay numeric")

	// Field order must be canonical.
	assert.Less(t, strings.Index(buf.String(), "name:"), strings.Index(buf.String(), "size:"))
}

func TestXMLROMs(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).XMLROMs(testROMs, []string{"name", "url"}))
	out := buf.String()

	var doc struct {
		ROMs []struct {
			Name string `xml:"name"`
			URL  string `xml:"url"`
		} `xml:"rom"`
	}
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &doc), "invalid XML:\n%s", out)

	require.Len(t, doc.ROMs, 2)
	assert.Equal(t, "Sonic & Knuckles (World)", doc.ROMs[0].Name)
	assert.True(t, strings.HasPrefix(out, xml.Header))
	assert.NotContains(t, out, "<hash>", "unselected fields must not be encoded")
}

func TestXMLSystems(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).XMLSystems(testSystems))

	var doc struct {
		Systems []ds.System `xml:"system"`
	}
	require.NoError(t, xml.Unmarshal(buf.Bytes(), &doc))
	assert.Equal(t, testSystems, doc.Systems)
}

func TestCSVROMs(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).CSVROMs(testROMs, []string{"name", "dls"}))

	rows, err := csv.NewReader(&buf).ReadAll()
	require.NoError(t, err)

	require.Len(t, rows, 3, "header plus two ROMs")
	assert.Equal(t, []string{"name", "dls"}, rows[0])
	assert.Equal(t, []string{"Sonic & Knuckles (World)", "341"}, rows[1],
		"names with commas and ampersands must survive CSV quoting")
}

func TestCSVROMsEmptyHasHeader(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).CSVROMs(nil, []string{"name", "url"}))

	assert.Equal(t, "name,url", strings.TrimSpace(buf.String()))
}

func TestCSVSystems(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(WithWriter(&buf)).CSVSystems(testSystems))

	rows, err := csv.NewReader(&buf).ReadAll()
	require.NoError(t, err)

	require.Len(t, rows, 3)
	assert.Equal(t, []string{"id", "name"}, rows[0])
	assert.Equal(t, []string{"atari-2600", "Atari 2600"}, rows[1])
}

func TestStructuredFormatsRejectUnknownColumn(t *testing.T) {
	p := New(WithWriter(&bytes.Buffer{}))

	require.Error(t, p.JSONROMs(testROMs, []string{"bogus"}))
	require.Error(t, p.YAMLROMs(testROMs, []string{"bogus"}))
	require.Error(t, p.XMLROMs(testROMs, []string{"bogus"}))
	require.Error(t, p.CSVROMs(testROMs, []string{"bogus"}))
}

func TestJSONAndYAMLGeneric(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf))

	require.NoError(t, p.JSON(testSystems))
	var js []ds.System
	require.NoError(t, json.Unmarshal(buf.Bytes(), &js))
	assert.Equal(t, testSystems, js)

	buf.Reset()
	require.NoError(t, p.YAML(testSystems))
	var ys []ds.System
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &ys))
	assert.Equal(t, testSystems, ys)
}

func TestWriterExposesDestination(t *testing.T) {
	var buf bytes.Buffer
	assert.Same(t, any(&buf), any(New(WithWriter(&buf)).Writer()))
}

func TestDefaultWriterIsStdout(t *testing.T) {
	assert.NotNil(t, New().Writer())
}
