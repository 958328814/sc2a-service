package dist

import (
	"testing"

	"reflect"

	"bytes"

	"sync"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestcreateLinks(t *testing.T) {
	subs := []string{"a", "b", "c"}
	links, err := createLinks(subs, "RELEASE_ID")
	assert.NoError(t, err)
	assert.Equal(t, len(subs), len(links))
	assert.NoError(t, clearLinks())
}

func TestGetLink(t *testing.T) {
	links, err := createLinks([]string{"a"}, "RELEASE_ID")
	assert.NoError(t, err)
	id := links[0].ID
	link, err := GetLink(id)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(link, &links[0]))
}

func TestStreamLink(t *testing.T) {
	buf := bytes.NewBuffer([]byte("OK"))
	r, err := Publish(Release{}, buf)
	assert.NoError(t, err)
	links, err := createLinks([]string{"a"}, r.ID)
	assert.NoError(t, err)
	id := links[0].ID
	assert.NoError(t, StreamLink(id, buf))
	assert.Equal(t, "OK", string(buf.Bytes()))

	assert.NoError(t, db.View(func(tx *bolt.Tx) (err error) {
		v := tx.Bucket([]byte("sub_download")).Bucket([]byte("a")).Get([]byte(r.ID))
		assert.Equal(t, v, []byte("1"))
		return
	}))
	assert.NoError(t, Unpublish(r.ID))
}

func TestGetSubDowloadStats(t *testing.T) {
	buf := bytes.NewBuffer([]byte("OK"))
	r, err := Publish(Release{}, buf)
	assert.NoError(t, err)
	links, err := createLinks([]string{"b"}, r.ID)
	assert.NoError(t, err)
	id := links[0].ID

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			assert.NoError(t, StreamLink(id, buf))
			wg.Done()
		}()
	}
	wg.Wait()

	stats, err := GetSubDowloadStats("b")
	assert.NoError(t, err)

	assert.Equal(t, uint64(10), stats[r.ID])

	assert.NoError(t, Unpublish(r.ID))
}
