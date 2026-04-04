package valobj

import (
	"strconv"
	"time"
)

type Header struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type QueryParam struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

type HTTPMethod string

const (
	MethodGet     HTTPMethod = "GET"
	MethodPost    HTTPMethod = "POST"
	MethodPut     HTTPMethod = "PUT"
	MethodDelete  HTTPMethod = "DELETE"
	MethodPatch   HTTPMethod = "PATCH"
	MethodHead    HTTPMethod = "HEAD"
	MethodOptions HTTPMethod = "OPTIONS"
)

func (m HTTPMethod) String() string {
	return string(m)
}

type Timing struct {
	DNSLookup    time.Duration `json:"dns_lookup"`
	TCPConnect   time.Duration `json:"tcp_connect"`
	TLSHandshake time.Duration `json:"tls_handshake"`
	TTFB         time.Duration `json:"ttfb"`
	Total        time.Duration `json:"total"`
}

func (t Timing) String() string {
	return strconv.Itoa(int(t.Total.Milliseconds())) + "ms"
}
