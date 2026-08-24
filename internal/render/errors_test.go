package render

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// failWriter fails every write, exercising the error paths of renderers.
type failWriter struct{}

var errBoom = errors.New("boom")

func (failWriter) Write([]byte) (int, error) { return 0, errBoom }

func TestRenderersSurfaceWriteErrors(t *testing.T) {
	p := New(WithWriter(failWriter{}))
	cols := []string{"name", "dls", "url"}

	require.Error(t, p.PrintROMs(testROMs, cols))
	require.Error(t, p.PrintSystems(testSystems))
	require.Error(t, p.JSON(testSystems))
	require.Error(t, p.YAML(testSystems))
	require.Error(t, p.JSONROMs(testROMs, cols))
	require.Error(t, p.YAMLROMs(testROMs, cols))
	require.Error(t, p.XMLROMs(testROMs, cols))
	require.Error(t, p.XMLSystems(testSystems))
	require.Error(t, p.CSVROMs(testROMs, cols))
	require.Error(t, p.CSVSystems(testSystems))
}
