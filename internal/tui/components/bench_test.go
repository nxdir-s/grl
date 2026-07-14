package components

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/nxdir-s/grl/internal/core/domain"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

const (
	BenchMediumJSON int = 10 * 1024
	BenchLargeJSON  int = 1024 * 1024
)

// benchAdapter implements the ports.TUI methods components touch, delegating
// JSON colorizing to the real formatter so benchmarks measure real work
type benchAdapter struct {
	ports.TUI
	formatter ports.Formatter

	colorizeCalls int
}

func newBenchAdapter() *benchAdapter {
	return &benchAdapter{
		formatter: domain.NewFormatter(),
	}
}

func (a *benchAdapter) ColorizeJSON(s string) string {
	a.colorizeCalls++
	return a.formatter.ColorizeJSON(s)
}

// makeBenchJSON returns a valid JSON document of approximately size bytes
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

func makeBenchResponse(size int) *entity.Response {
	body := makeBenchJSON(size)

	return &entity.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []valobj.Header{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "Cache-Control", Value: "no-cache"},
			{Key: "X-Request-Id", Value: "bench-request-id"},
		},
		Body:          bytes.NewBufferString(body),
		ContentType:   "application/json",
		ContentLength: int64(len(body)),
	}
}

func BenchmarkKVEditorView(b *testing.B) {
	for _, n := range []int{20, 100} {
		b.Run(fmt.Sprintf("rows_%d", n), func(b *testing.B) {
			headers := make([]valobj.Header, 0, n)
			for i := 0; i < n; i++ {
				headers = append(headers, valobj.Header{
					Key:     fmt.Sprintf("X-Header-%d", i),
					Value:   strings.Repeat("v", 24),
					Enabled: i%2 == 0,
				})
			}

			editor := NewKVEditor("key", "value")
			editor.SetHeaders(headers)
			editor.SetWidth(80)

			b.ReportAllocs()

			for b.Loop() {
				_ = editor.View()
			}
		})
	}
}

func BenchmarkWrapForViewport(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"10KB", BenchMediumJSON},
		{"1MB", BenchLargeJSON},
	}

	for _, bc := range sizes {
		b.Run(fmt.Sprintf("size_%s", bc.name), func(b *testing.B) {
			pane := NewResponseBodyPane(newBenchAdapter())
			content := makeBenchResponse(bc.size).FormatBody()

			b.ReportAllocs()
			b.SetBytes(int64(len(content)))

			for b.Loop() {
				_ = pane.wrapForViewport(content, 80)
			}
		})
	}
}

// BenchmarkSearchKeystroke measures the cost of one typed and one deleted
// character in the response search bar with a large body loaded
func BenchmarkSearchKeystroke(b *testing.B) {
	viewer := NewResponseViewer(newBenchAdapter())
	viewer.SetSize(100, 40)
	viewer.SetResponse(makeBenchResponse(BenchLargeJSON))
	viewer.Focus()

	if cmd := viewer.OpenSearch(); cmd == nil {
		b.Fatal("expected search to open")
	}

	typeKey := tea.KeyPressMsg{Code: 'a', Text: "a"}
	deleteKey := tea.KeyPressMsg{Code: tea.KeyBackspace}

	b.ReportAllocs()

	for b.Loop() {
		viewer, _ = viewer.Update(typeKey)
		viewer, _ = viewer.Update(deleteKey)
	}
}

// BenchmarkDragResize measures an interactive drag-resize: a stream of
// width changes followed by the settled flush that refreshes the body pane
func BenchmarkDragResize(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"10KB", BenchMediumJSON},
		{"1MB", BenchLargeJSON},
	}

	widths := make([]int, 30)
	for i := range widths {
		widths[i] = 80 + i
	}

	for _, bc := range sizes {
		b.Run(fmt.Sprintf("size_%s", bc.name), func(b *testing.B) {
			viewer := NewResponseViewer(newBenchAdapter())
			viewer.SetSize(120, 40)
			viewer.SetResponse(makeBenchResponse(bc.size))

			b.ReportAllocs()

			for b.Loop() {
				for _, w := range widths {
					viewer.SetSize(w, 40)
				}

				viewer.FlushResize()
			}
		})
	}
}
