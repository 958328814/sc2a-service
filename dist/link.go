package dist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"strconv"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
)

// Link is a public link send to subscriber
type Link struct {
	ID        string
	SubID     string
	ReleaseID string
	Date      time.Time
}

func createLinks(subs []string, releaseID string) (rv []Link, err error) {
	now := time.Now()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("link"))
		for _, sub := range subs {
			id := uuid.NewV4().String()
			link := Link{
				ID:        id,
				SubID:     sub,
				ReleaseID: releaseID,
				Date:      now,
			}
			rv = append(rv, link)
			j, err := json.Marshal(link)
			if err != nil {
				return err
			}
			err = b.Put([]byte(id), j)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func removeLinkIf(fn func(link *Link) bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("link"))
		b.ForEach(func(k, v []byte) error {
			link := Link{}
			if err := json.Unmarshal(v, &link); err != nil {
				return fmt.Errorf("link unmarshal: %s", err.Error())
			}
			if fn(&link) {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
			return nil
		})
		return nil
	})
}

func removeSubLinks(subID string) error {
	return removeLinkIf(func(link *Link) bool {
		return link.SubID == subID
	})
}

func removeReleaseLinks(releaseID string) error {
	return removeLinkIf(func(link *Link) bool {
		return link.ReleaseID == releaseID
	})
}

func clearLinks() error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("link"))
	})
}

// ErrLinkNotFound is returned when a link id is not found
var ErrLinkNotFound = errors.New("link was not found")

// GetLink returns link by ID
func GetLink(id string) (rv *Link, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket([]byte("link")).Get([]byte(id))
		if v == nil {
			return ErrLinkNotFound
		}
		link := Link{}
		if err := json.Unmarshal(v, &link); err != nil {
			return err
		}
		rv = &link
		return nil
	})
	return
}

var downloadLock sync.RWMutex

// StreamLink streams release data to a writer and increases download count
func StreamLink(id string, w io.Writer) (err error) {
	downloadLock.Lock()
	defer downloadLock.Unlock()
	link, err := GetLink(id)
	if err != nil {
		return
	}
	err = db.Update(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte("sub_download")).Bucket([]byte(link.SubID))
		if b == nil {
			b, err = tx.Bucket([]byte("sub_download")).CreateBucketIfNotExists([]byte(link.SubID))
			if err != nil {
				return
			}
		}

		v := b.Get([]byte(link.ReleaseID))
		var n uint64
		if v != nil {
			n, _ = strconv.ParseUint(string(v), 16, 64)
		}
		n = n + 1
		err = b.Put([]byte(link.ReleaseID), []byte(strconv.FormatUint(n, 16)))
		return
	})

	err = Stream(link.ReleaseID, w)
	return
}

// GetSubDowloadStats returns download statistic for a subscriber
func GetSubDowloadStats(subID string) (rv map[string]uint64, err error) {
	downloadLock.RLock()
	defer downloadLock.RUnlock()

	rv = map[string]uint64{}
	err = db.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte("sub_download")).Bucket([]byte(subID))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				var n uint64
				n, _ = strconv.ParseUint(string(v), 16, 64)
				rv[string(k)] = n
				return nil
			})
		}
		return nil
	})
	return
}
