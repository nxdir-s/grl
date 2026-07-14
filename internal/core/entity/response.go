package entity

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

const (
	JSONContentType string = "json"
)

type Response struct {
	StatusCode int             `json:"status_code"`
	Status     string          `json:"status"`
	Headers    []valobj.Header `json:"headers,omitempty"`
	// Body is not persisted: it always marshaled as an empty object, and
	// storing full bodies would bloat the history file
	Body          *bytes.Buffer `json:"-"`
	ContentType   string        `json:"content_type"`
	ContentLength int64         `json:"content_length"`
	Timing        valobj.Timing `json:"timing"`
}

func NewResponse() *Response {
	return &Response{}
}

func (e *Response) FormatBody() string {
	if e.Body == nil || len(e.Body.Bytes()) == 0 {
		return ""
	}

	if strings.Contains(e.ContentType, JSONContentType) || e.looksLikeJSON() {
		pretty := &bytes.Buffer{}
		if err := json.Indent(pretty, e.Body.Bytes(), "", "  "); err == nil {
			return pretty.String()
		}
	}

	return e.Body.String()
}

func (e *Response) looksLikeJSON() bool {
	trimmed := bytes.TrimSpace(e.Body.Bytes())
	if len(trimmed) == 0 {
		return false
	}

	return (trimmed[0] == '{' || trimmed[0] == '[') &&
		(trimmed[len(trimmed)-1] == '}' || trimmed[len(trimmed)-1] == ']')
}
