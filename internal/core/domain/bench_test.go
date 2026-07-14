package domain

import (
	"fmt"
	"strings"
)

const (
	BenchMediumJSON int = 10 * 1024
	BenchLargeJSON  int = 1024 * 1024
)

// makeJSON returns a valid JSON document of approximately size bytes with a
// realistic mix of keys, strings, numbers, bools, and nulls
func makeJSON(size int) string {
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

// makeJSONWithVars returns a valid JSON document of approximately size bytes
// where each element contains {{host}} and {{token}} placeholders
func makeJSONWithVars(size int) string {
	var sb strings.Builder
	sb.Grow(size + 128)

	sb.WriteString(`{"items":[`)

	for i := 0; sb.Len() < size-16; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}

		fmt.Fprintf(&sb, `{"id":%d,"endpoint":"https://{{host}}/items/%d","auth":"Bearer {{token}}"}`, i, i)
	}

	sb.WriteString(`]}`)

	return sb.String()
}
