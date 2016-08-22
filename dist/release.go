package dist

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

var idFormat = "20060102T150405Z0700"
var publishMutex sync.RWMutex

// Release represents a published version
type Release struct {
	ID          string
	Version     string
	Description string
	Date        time.Time
}

// DataDir is the path to store files
const DataDir = "./data"

func init() {
	err := os.MkdirAll(DataDir, 0400)
	if err != nil {
		log.Fatal(err)
	}
}

func openReleaseFile() (string, *os.File, error) {
	ts := time.Now().Format(idFormat)
	tc := 0
	for {
		var name string
		if tc == 0 {
			name = ts
		} else {
			name = fmt.Sprintf("%s_%d", ts, tc)
		}

		f, err := os.OpenFile(dataFilePath(name+".json"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0400)
		if err != nil {
			if os.IsExist(err) {
				tc = tc + 1
				continue
			}
			return "", nil, err
		}
		return name, f, nil
	}
}

func dataFilePath(name string) string {
	return DataDir + "/" + name
}

// List list all releases
func List() ([]Release, error) {
	publishMutex.RLock()
	defer publishMutex.RUnlock()

	files, err := ioutil.ReadDir(DataDir)
	if err != nil {
		return nil, err
	}

	list := []Release{}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			f, err := os.Open(dataFilePath(file.Name()))
			if err != nil {
				return nil, err
			}
			item := Release{}
			err = json.NewDecoder(f).Decode(&item)
			if err != nil {
				return nil, fmt.Errorf("unable to decode file: %s", file.Name())
			}
			list = append(list, item)
		}
	}

	return list, nil
}

// Publish uploads & publishes a new version
func Publish(release Release, r io.Reader) (*Release, error) {
	publishMutex.Lock()
	defer publishMutex.Unlock()

	id, f, err := openReleaseFile()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	saved := &Release{
		ID:          id,
		Version:     release.Version,
		Description: release.Description,
		Date:        time.Now(),
	}

	dataf, err := os.OpenFile(dataFilePath(id+".dat"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 400)
	if err != nil {
		os.RemoveAll(f.Name())
		return nil, err
	}
	defer dataf.Close()

	_, err = io.Copy(dataf, r)
	if err != nil {
		os.RemoveAll(f.Name())
		return nil, err
	}

	err = json.NewEncoder(f).Encode(saved)
	if err != nil {
		return nil, err
	}

	return saved, nil
}

func checkID(id string) error {
	if id == "" {
		return fmt.Errorf("ID is required")
	}

	if len(id) != len(idFormat) {
		return fmt.Errorf("%v is not a valid ID", id)
	}

	base := path.Clean(DataDir)
	p := path.Clean(dataFilePath(id))
	if path.Dir(p) != base {
		return fmt.Errorf("%v is not an illegal ID", id)
	}
	return nil
}

// Unpublish deletes a published version
func Unpublish(id string) error {
	if err := checkID(id); err != nil {
		return err
	}

	publishMutex.Lock()
	defer publishMutex.Unlock()

	if err := os.RemoveAll(dataFilePath(id + ".dat")); err != nil {
		return err
	}

	if err := os.RemoveAll(dataFilePath(id + ".json")); err != nil {
		return err
	}

	return nil
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
	if err := checkID(id); err != nil {
		return nil, err
	}

	publishMutex.RLock()
	defer publishMutex.RUnlock()

	f, err := getDataReader(id + ".json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := Release{}
	err = json.NewDecoder(f).Decode(&r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Stream streams release file data to a writer
func Stream(id string, w io.Writer) error {
	if err := checkID(id); err != nil {
		return err
	}

	publishMutex.RLock()
	defer publishMutex.RUnlock()

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