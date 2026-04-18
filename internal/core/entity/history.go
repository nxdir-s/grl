package entity

import (
	"time"

	"github.com/google/uuid"
)

type History struct{}

func NewHistory() *History {
	return &History{}
}

type HistoryEntry struct {
	ID        string    `json:"id"`
	Request   *Request  `json:"request"`
	Response  *Response `json:"response,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHistoryEntry(req *Request, resp *Response) HistoryEntry {
	return HistoryEntry{
		ID:        uuid.New().String(),
		Request:   req,
		Response:  resp,
		Timestamp: time.Now(),
	}
}
