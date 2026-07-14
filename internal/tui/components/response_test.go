package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func typeSearch(t *testing.T, viewer ResponseViewer, pattern string) ResponseViewer {
	t.Helper()

	for _, r := range pattern {
		viewer, _ = viewer.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}

	return viewer
}

func TestResizeDebounceColorizesOnce(t *testing.T) {
	adapter := newBenchAdapter()

	viewer := NewResponseViewer(adapter)
	viewer.SetSize(120, 40)
	viewer.SetResponse(makeBenchResponse(BenchMediumJSON))

	adapter.colorizeCalls = 0

	for w := 80; w < 110; w++ {
		viewer.SetSize(w, 40)
	}

	assert.Equal(t, 0, adapter.colorizeCalls, "resize stream must not re-colorize per width")

	viewer.FlushResize()
	assert.Equal(t, 1, adapter.colorizeCalls, "settled resize must re-colorize exactly once")

	viewer.FlushResize()
	assert.Equal(t, 1, adapter.colorizeCalls, "flush without a pending resize must be a no-op")
}

func TestResizeHeightOnlyDoesNotRefresh(t *testing.T) {
	adapter := newBenchAdapter()

	viewer := NewResponseViewer(adapter)
	viewer.SetSize(120, 40)
	viewer.SetResponse(makeBenchResponse(BenchMediumJSON))

	adapter.colorizeCalls = 0

	viewer.SetSize(120, 30)
	viewer.FlushResize()

	assert.Equal(t, 0, adapter.colorizeCalls, "height-only resize must not re-colorize")
}

func TestSearchMatchesBody(t *testing.T) {
	adapter := newBenchAdapter()

	resp := makeBenchResponse(BenchMediumJSON)
	expected := strings.Count(strings.ToLower(resp.FormatBody()), "item")

	viewer := NewResponseViewer(adapter)
	viewer.SetSize(120, 40)
	viewer.SetResponse(resp)
	viewer.Focus()

	assert.NotNil(t, viewer.OpenSearch())

	// mixed case exercises the lowercased haystack cache
	viewer = typeSearch(t, viewer, "ITem")

	assert.Equal(t, expected, viewer.search.total)
	assert.Equal(t, 1, viewer.search.current)

	// deleting a character re-runs the match against the cached haystack
	viewer, _ = viewer.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	assert.Equal(t, strings.Count(strings.ToLower(resp.FormatBody()), "ite"), viewer.search.total)
}

func TestSearchMatchesHeadersAfterTabSwitch(t *testing.T) {
	adapter := newBenchAdapter()

	resp := makeBenchResponse(BenchMediumJSON)

	viewer := NewResponseViewer(adapter)
	viewer.SetSize(120, 40)
	viewer.SetResponse(resp)
	viewer.Focus()

	assert.NotNil(t, viewer.OpenSearch())

	viewer = typeSearch(t, viewer, "content")

	// switch to the headers tab mid-search; matches must recompute against
	// the header pane's cached haystack
	viewer, _ = viewer.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})

	expected := strings.Count(strings.ToLower(viewer.headerPane.Raw()), "content")
	assert.Positive(t, expected)
	assert.Equal(t, expected, viewer.search.total)
}

func TestSearchClearedOnNewResponse(t *testing.T) {
	adapter := newBenchAdapter()

	viewer := NewResponseViewer(adapter)
	viewer.SetSize(120, 40)
	viewer.SetResponse(makeBenchResponse(BenchMediumJSON))
	viewer.Focus()

	assert.NotNil(t, viewer.OpenSearch())
	viewer = typeSearch(t, viewer, "item")
	assert.Positive(t, viewer.search.total)

	viewer.Clear()

	assert.Empty(t, viewer.bodyPane.Raw())
	assert.Empty(t, viewer.bodyPane.RawLower())
	assert.False(t, viewer.SearchActive())
}
