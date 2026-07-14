package components

import (
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

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
	RequestName  string
}

type CreateAndSaveMsg struct {
	CollectionName string
	RequestName    string
}

type CloseSaveModalMsg struct{}

type RequestSavedMsg struct {
	Collections []entity.Collection
	FlashText   string
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

type ConfigUpdatedMsg struct {
	Cfg *valobj.Config
}

type RenameCollectionMsg struct {
	ID      string
	NewName string
}

type DeleteCollectionMsg struct {
	ID string
}

type RemoveRequestMsg struct {
	CollectionID string
	RequestID    string
}

type DeleteHistoryEntryMsg struct {
	EntryID string
}

// ResizeSettledMsg fires after a resize stream has been quiet long enough to
// apply deferred refreshes; Seq identifies the resize event that scheduled it
type ResizeSettledMsg struct {
	Seq int
}
