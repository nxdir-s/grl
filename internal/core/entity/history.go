package entity

import "time"

type History struct{}

func NewHistory() *History {
	return &History{}
}

type HistoryEntry struct {
	Request   *Request  `json:"request"`
	Response  *Response `json:"response,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHistoryEntry(req *Request, resp *Response) *HistoryEntry {
	return &HistoryEntry{
		Request:   req,
		Response:  resp,
		Timestamp: time.Now(),
	}
}
