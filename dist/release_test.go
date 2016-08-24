package dist

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	now := time.Now()
	testdata := []Release{
		Release{Version: "0.01", Description: "D1", Date: now},
		Release{Version: "0.02", Description: "D2", Date: now.Add(time.Second)},
		Release{Version: "0.03", Description: "D3", Date: now.Add(time.Second * 2)},
	}

	err := db.Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("release"))

		for _, r := range testdata {
			j, err := json.Marshal(r)
			if err != nil {
				return err
			}
			seq, _ := b.NextSequence()
			err = b.Put([]byte(strconv.FormatUint(seq, 10)), j)
			if err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)
	rl, err := List()
	assert.NoError(t, err)

	sort.Stable(ReleasesByDateDesc(testdata))

	assert.Equal(t, len(testdata), len(rl))
	for i := range testdata {
		assert.True(t, reflect.DeepEqual(testdata[i], rl[i]))
	}
}

func testGetRelease(id string) (r *Release, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket([]byte("release")).Get([]byte(id))
		if v == nil {
			return nil
		}
		fromDb := Release{}
		err = json.Unmarshal(v, &fromDb)
		if err != nil {
			return err
		}
		r = &fromDb
		return nil
	})
	return
}

func TestPublish(t *testing.T) {
	buf := bytes.NewBufferString("RELEASE")
	r, err := Publish(Release{
		Version:     "0.0.1",
		Description: "Test publish",
	}, buf)

	assert.NoError(t, err)
	assert.NotNil(t, r)

	fpath := dataFilePath(r.ID + ".dat")
	fdata, err := os.Open(fpath)
	assert.NoError(t, err)
	c, err := ioutil.ReadAll(fdata)
	assert.NoError(t, err)
	assert.Equal(t, "RELEASE", string(c))
	fdata.Close()

	fromDb, err := testGetRelease(r.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fromDb)

	assert.Equal(t, r.Version, fromDb.Version)
	assert.Equal(t, r.Description, fromDb.Description)

	assert.NoError(t, db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("release")).Delete([]byte(r.ID))
	}))
	assert.NoError(t, os.RemoveAll(fpath))
}

func TestUnpublish(t *testing.T) {
	created := Release{
		Version:     "0.0.1",
		Description: "Test publish",
	}
	buf := bytes.NewBufferString("RELEASE")
	r, err := Publish(created, buf)
	assert.NoError(t, err)
	assert.NoError(t, Unpublish(r.ID))

	_, err = os.Stat(dataFilePath(r.ID + ".dat"))
	assert.True(t, os.IsNotExist(err))

	r, err = testGetRelease(r.ID)
	assert.NoError(t, err)
	assert.Equal(t, (*Release)(nil), r)
}

func TestGet(t *testing.T) {
	created := Release{
		Version:     "0.0.1",
		Description: "Test publish",
	}
	buf := bytes.NewBufferString("RELEASE")
	r, err := Publish(created, buf)
	assert.NoError(t, err)

	fromJSON, err := Get(r.ID)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(fromJSON, r))
	assert.NoError(t, Unpublish(r.ID))
}

func TestStream(t *testing.T) {
	created := Release{
		Version:     "0.0.1",
		Description: "Test publish",
	}
	buf := bytes.NewBufferString("RELEASE")
	r, err := Publish(created, buf)
	assert.NoError(t, err)

	outBuf := bytes.NewBuffer(nil)
	err = Stream(r.ID, outBuf)
	assert.NoError(t, err)
	assert.Equal(t, "RELEASE", string(outBuf.Bytes()))
	assert.NoError(t, Unpublish(r.ID))
}
