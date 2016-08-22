package dist

import "sync"

var subsLock sync.RWMutex
var subListFile = dataFilePath("subs.json")

// Sub represents a email subscriber
type Sub struct {
	ID    string
	Name  string
	Email string
	Date  string
}

// ListSubs returns subscriber list
func ListSubs() ([]Sub, error) {
	return nil, nil
}

// Subscribe adds a new subscriber
func Subscribe(sub Sub) (*Sub, error) {
	return nil, nil
}

// Unsubscribe removes a subscriber by ID
func Unsubscribe(id string) error {
	return nil
}
