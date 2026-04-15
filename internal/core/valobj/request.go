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

func (s HTTPMethod) String() string {
	return string(s)
}

type Timing struct {
	DNSLookup    time.Duration `json:"dns_lookup"`
	TCPConnect   time.Duration `json:"tcp_connect"`
	TLSHandshake time.Duration `json:"tls_handshake"`
	TTFB         time.Duration `json:"ttfb"`
	Total        time.Duration `json:"total"`
}

func (s Timing) String() string {
	return strconv.Itoa(int(s.Total.Milliseconds())) + "ms"
}
