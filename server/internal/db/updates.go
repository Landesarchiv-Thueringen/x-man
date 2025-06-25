package db

// Update notifies interested parties of updates in the database.
//
// Not all collections broadcast updates. Add broadcast functions for
// collections as needed.
type Update struct {
	// Collection is the database collection that has been updated.
	Collection string `json:"collection"`
	// ProcessID is set if the update can be attributed to one single process.
	ProcessID *string `json:"processId"`
	// Operation is the kind of database operation that ocurred.
	Operation UpdateOperation `json:"operation"`
}

type UpdateOperation string

const (
	UpdateOperationInsert UpdateOperation = "insert"
	UpdateOperationUpdate UpdateOperation = "update"
	UpdateOperationDelete UpdateOperation = "delete"
)

var updatesChannels []chan Update

func RegisterUpdatesChannel() chan Update {
	ch := make(chan Update)
	updatesChannels = append(updatesChannels, ch)
	return ch
}

func UnregisterUpdatesChannel(ch chan Update) {
	for i, c := range updatesChannels {
		if c == ch {
			updatesChannels[i] = updatesChannels[len(updatesChannels)-1]
			updatesChannels = updatesChannels[:len(updatesChannels)-1]
			return
		}
	}
	panic("updates channel not found")
}

func broadcastUpdate(u Update) {
	for _, ch := range updatesChannels {
		ch <- u
	}
}
