package secondary

import (
	"bytes"
	"io"
	"log/slog"
)

const (
	BenchSmallBody  int64 = 50
	BenchMediumBody int64 = 10 * Kib
	BenchLargeBody  int64 = 1 * Mib
)

// benchLogger returns a logger that discards output so benchmarks don't measure log formatting
func benchLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// makeJSONPayload returns a JSON document of approximately size bytes
func makeJSONPayload(size int64) []byte {
	const wrapper = `{"data":""}`

	fill := size - int64(len(wrapper))
	if fill < 0 {
		fill = 0
	}

	payload := make([]byte, 0, int64(len(wrapper))+fill)
	payload = append(payload, `{"data":"`...)
	payload = append(payload, bytes.Repeat([]byte("x"), int(fill))...)
	payload = append(payload, `"}`...)

	return payload
}
