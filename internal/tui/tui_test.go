package tui

import (
	"bytes"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/nxdir-s/grl/internal/core/domain"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/tui/components"
)

func TestResizeDebounceSeqGating(t *testing.T) {
	adapter := &benchAdapter{formatter: domain.NewFormatter()}

	body := makeBenchJSON(10 * 1024)
	resp := &entity.Response{
		StatusCode:  200,
		Status:      "200 OK",
		Body:        bytes.NewBufferString(body),
		ContentType: "application/json",
	}

	model := New(adapter)
	model.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	model.Update(components.ResponseReceivedMsg{
		Response: resp,
		Request:  &entity.Request{Method: valobj.MethodGet, URL: "https://example.com"},
	})

	adapter.colorizeCalls = 0

	model.Update(tea.WindowSizeMsg{Width: 180, Height: 60})
	model.Update(tea.WindowSizeMsg{Width: 190, Height: 60})

	assert.Equal(t, 0, adapter.colorizeCalls, "resize stream must not re-colorize per event")

	model.Update(components.ResizeSettledMsg{Seq: model.resizeSeq - 1})
	assert.Equal(t, 0, adapter.colorizeCalls, "stale resize tick must be dropped")

	model.Update(components.ResizeSettledMsg{Seq: model.resizeSeq})
	assert.Equal(t, 1, adapter.colorizeCalls, "settled resize must re-colorize exactly once")
}
