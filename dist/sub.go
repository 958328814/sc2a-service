package dist

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"errors"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
)

// Sub represents a email subscriber
type Sub struct {
	ID    string
	Name  string
	Email string
	Date  time.Time
}

// SubsByDate is slice of Release sorted by date
type SubsByDate []Sub

func (l SubsByDate) Len() int           { return len(l) }
func (l SubsByDate) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l SubsByDate) Less(i, j int) bool { return l[i].Date.Before(l[j].Date) }

// ListSubs returns subscriber list
func ListSubs() (SubsByDate, error) {
	list := SubsByDate{}

	err := db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("sub")).ForEach(func(k, v []byte) error {
			s := Sub{}
			err := json.Unmarshal(v, &s)
			if err != nil {
				return fmt.Errorf("unmarshal subscriber %s: %s", string(k), err.Error())
			}
			list = append(list, s)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	sort.Stable(list)

	return list, nil
}

// Subscribe adds a new subscriber
func Subscribe(sub Sub) (rv *Sub, err error) {
	id := uuid.NewV4().String()
	sub.ID = id
	sub.Date = time.Now()
	j, err := json.Marshal(sub)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("sub"))
		return b.Put([]byte(id), j)
	})

	if err != nil {
		return
	}

	rv = &sub
	return
}

// ErrSubNotFound is returned when subscriber is not found by id
var ErrSubNotFound = errors.New("subscriber was not found")

// UpdateSubscriber updates subscribes' email and name
func UpdateSubscriber(sub Sub) (rv *Sub, err error) {
	id := sub.ID
	fromDb := Sub{}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("sub"))
		v := b.Get([]byte(id))

		if v == nil {
			return ErrSubNotFound
		}

		err := json.Unmarshal(v, &fromDb)
		if err != nil {
			return err
		}

		fromDb.Email = sub.Email
		fromDb.Name = sub.Name

		j, err := json.Marshal(fromDb)
		if err != nil {
			return err
		}

		return b.Put([]byte(id), j)
	})
	if err != nil {
		return
	}
	return &fromDb, nil
}

// Unsubscribe removes a subscriber by ID
func Unsubscribe(id string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("sub")).Delete([]byte(id))
	})

	if err != nil {
		return err
	}

	return nil
}
