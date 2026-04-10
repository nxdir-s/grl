package components

import "github.com/nxdir-s/grl/internal/core/entity"

type ResponseReceivedMsg struct {
	Response *entity.Response
	Request  *entity.Request
}

type RequestErrorMsg struct {
	Err error
}

type SendRequestMsg struct{}

type HistoryUpdatedMsg struct {
	History []entity.HistoryEntry
}

type CollectionsUpdatedMsg struct {
	Collections []entity.Collection
}

type SaveToCollectionMsg struct {
	CollectionID string
}

type ErrorMsg struct {
	Err error
}

type EnvironmentsUpdatedMsg struct {
	Environments []entity.Environment
	Active       *entity.Environment
}

type EnvironmentSwitchedMsg struct {
	Active *entity.Environment
}

type FlashMsg struct {
	Text string
}

type ClearFlashMsg struct{}
