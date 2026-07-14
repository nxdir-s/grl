package tui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/nxdir-s/grl/internal/core/domain"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
	"github.com/nxdir-s/grl/internal/tui/components"
)

// benchAdapter implements the ports.TUI methods the render path touches,
// delegating JSON colorizing to the real formatter
type benchAdapter struct {
	ports.TUI
	formatter   ports.Formatter
	collections []entity.Collection

	colorizeCalls int
}

func (a *benchAdapter) ColorizeJSON(s string) string {
	a.colorizeCalls++
	return a.formatter.ColorizeJSON(s)
}

func (a *benchAdapter) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	return a.collections, nil
}

func makeBenchJSON(size int) string {
	var sb strings.Builder
	sb.Grow(size + 128)

	sb.WriteString(`{"items":[`)

	for i := 0; sb.Len() < size-16; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}

		fmt.Fprintf(&sb, `{"id":%d,"name":"item-%d","active":true,"score":98.6,"tags":["alpha","beta"],"note":null}`, i, i)
	}

	sb.WriteString(`]}`)

	return sb.String()
}

// BenchmarkTUIView measures a full frame render of the root model at 200x60
// with a populated sidebar and a ~100KB response loaded
func BenchmarkTUIView(b *testing.B) {
	collections := make([]entity.Collection, 0, 5)
	for i := 0; i < 5; i++ {
		requests := make([]entity.Request, 0, 10)
		for j := 0; j < 10; j++ {
			requests = append(requests, entity.Request{
				ID:     fmt.Sprintf("req-%d-%d", i, j),
				Name:   fmt.Sprintf("request %d", j),
				Method: valobj.MethodGet,
				URL:    fmt.Sprintf("https://api.example.com/v%d/resource/%d", i, j),
			})
		}

		collections = append(collections, entity.Collection{
			ID:       fmt.Sprintf("col-%d", i),
			Name:     fmt.Sprintf("collection %d", i),
			Requests: requests,
		})
	}

	history := make([]entity.HistoryEntry, 0, 50)
	for i := 0; i < 50; i++ {
		history = append(history, entity.HistoryEntry{
			ID: fmt.Sprintf("hist-%d", i),
			Request: &entity.Request{
				Method: valobj.MethodGet,
				URL:    fmt.Sprintf("https://api.example.com/users/%d", i),
			},
			Timestamp: time.Now(),
		})
	}

	adapter := &benchAdapter{
		formatter:   domain.NewFormatter(),
		collections: collections,
	}

	body := makeBenchJSON(100 * 1024)
	resp := &entity.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []valobj.Header{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "Cache-Control", Value: "no-cache"},
		},
		Body:          bytes.NewBufferString(body),
		ContentType:   "application/json",
		ContentLength: int64(len(body)),
		Timing: valobj.Timing{
			Total: 123 * time.Millisecond,
		},
	}

	model := New(adapter)

	model.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	model.Update(components.HistoryUpdatedMsg{History: history})
	model.Update(components.EnvironmentsUpdatedMsg{
		Environments: []entity.Environment{{ID: "env-1", Name: "staging"}},
		Active:       &entity.Environment{ID: "env-1", Name: "staging"},
	})
	model.Update(components.ResponseReceivedMsg{
		Response: resp,
		Request:  &entity.Request{Method: valobj.MethodGet, URL: "https://api.example.com"},
	})

	b.ReportAllocs()

	for b.Loop() {
		_ = model.View()
	}
}
