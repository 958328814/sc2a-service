package dist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"sort"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
)

// Release represents a published version
type Release struct {
	ID          string
	Version     string
	Description string
	Date        time.Time
}

// FileName generate file name of this release
func (r Release) FileName() string {
	type ctx struct {
		Version string
		Date    string
	}
	buf := bytes.NewBuffer(nil)
	nameTemplate.Execute(buf, ctx{
		Version: r.Version,
		Date:    r.Date.Format("20060102150405"),
	})
	return string(buf.Bytes())
}

// ReleasesByDateDesc is slice of Release sorted by date desc
type ReleasesByDateDesc []Release

func (rl ReleasesByDateDesc) Len() int           { return len(rl) }
func (rl ReleasesByDateDesc) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl ReleasesByDateDesc) Less(i, j int) bool { return !rl[i].Date.Before(rl[j].Date) }

// DataDir is the path to store files
const DataDir = "./data/release"

func dataFilePath(name string) string {
	return DataDir + "/" + name
}

func init() {
	err := os.MkdirAll(DataDir, 0400)
	if err != nil {
		log.Fatal(err)
	}
}

// List list all releases
func List() (ReleasesByDateDesc, error) {
	list := ReleasesByDateDesc{}

	err := db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("release")).ForEach(func(k, v []byte) error {
			r := Release{}
			err := json.Unmarshal(v, &r)
			if err != nil {
				return fmt.Errorf("unmarshal release %s: %s", string(k), err.Error())
			}
			list = append(list, r)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	sort.Stable(list)

	return list, nil
}

// Publish uploads & publishes a new version
func Publish(release Release, r io.Reader) (rv *Release, err error) {
	id := uuid.NewV4().String()
	saved := Release{
		ID:          id,
		Version:     release.Version,
		Description: release.Description,
		Date:        time.Now(),
	}

	fpath := dataFilePath(id + ".dat")
	dataf, err := os.OpenFile(fpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 400)
	if err != nil {
		return
	}
	defer func() {
		dataf.Close()
		if err != nil {
			os.RemoveAll(fpath)
		}
	}()

	_, err = io.Copy(dataf, r)
	if err != nil {
		return
	}

	j, err := json.Marshal(saved)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("release"))
		err = b.Put([]byte(id), j)
		return nil
	})

	rv = &saved
	return
}

func getRelease(id string) (r *Release, err error) {
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

// Unpublish deletes a published version
func Unpublish(id string) error {
	r, err := getRelease(id)
	if err != nil {
		return err
	}

	if r == nil {
		return nil
	}

	if err := os.RemoveAll(dataFilePath(id + ".dat")); err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("release")).Delete([]byte(id))
	})
}

func getDataReader(name string) (io.ReadCloser, error) {
	p := dataFilePath(name)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Get returns release metadata by id
func Get(id string) (*Release, error) {
	return getRelease(id)
}

// ErrReleaseDataNotFound is returned when a linked release data file is not found
var ErrReleaseDataNotFound = errors.New("release data file was not found")

// Stream streams release file data to a writer
func Stream(id string, w io.Writer) error {
	r, err := getRelease(id)
	if err != nil {
		return err
	}

	if r == nil {
		return ErrReleaseDataNotFound
	}

	f, err := getDataReader(id + ".dat")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	return nil
}
