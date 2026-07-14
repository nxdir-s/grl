package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func BenchmarkColorizeJSON(b *testing.B) {
	formatter := NewFormatter()

	sizes := []struct {
		name string
		size int
	}{
		{"10KB", BenchMediumJSON},
		{"1MB", BenchLargeJSON},
	}

	for _, bc := range sizes {
		// indent to mirror what the response pane feeds the formatter
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, []byte(makeJSON(bc.size)), "", "  "); err != nil {
			b.Fatalf("failed to indent payload: %s", err.Error())
		}

		input := pretty.String()

		b.Run(fmt.Sprintf("size_%s", bc.name), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))

			for b.Loop() {
				_ = formatter.ColorizeJSON(input)
			}
		})
	}
}
