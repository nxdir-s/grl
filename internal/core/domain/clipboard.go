package domain

import "github.com/atotto/clipboard"

type Clipboard struct{}

func NewClipboard() *Clipboard {
	return &Clipboard{}
}

func (d *Clipboard) Copy(s string) error {
	return clipboard.WriteAll(s)
}
