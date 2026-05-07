package entity

import (
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type RequestOpts func(e *Request)

func WithHeader(headers []valobj.Header) RequestOpts {
	return func(e *Request) {
		e.Headers = append(e.Headers, headers...)
	}
}

type Request struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Method      valobj.HTTPMethod   `json:"method"`
	URL         string              `json:"url"`
	Auth        valobj.Auth         `json:"auth,omitempty"`
	Headers     []valobj.Header     `json:"headers,omitempty"`
	Params      []valobj.QueryParam `json:"params,omitempty"`
	Body        string              `json:"body,omitempty"`
	BodyType    valobj.BodyType     `json:"body_type,omitempty"`
	FormFields  []valobj.FormField  `json:"form_fields,omitempty"`
	ContentType string              `json:"content_type,omitempty"`
}

func NewRequest(opts ...RequestOpts) *Request {
	return &Request{
		Headers: make([]valobj.Header, 0),
		Params:  make([]valobj.QueryParam, 0),
	}
}
