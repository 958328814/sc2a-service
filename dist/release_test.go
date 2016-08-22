package dist

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"encoding/json"

	"reflect"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	assert.NoError(t, ioutil.WriteFile(DataDir+"/1.json", []byte("{}"), 0755))
	assert.NoError(t, ioutil.WriteFile(DataDir+"/1.dat", []byte("1"), 0755))

	_, err := List()
	assert.NoError(t, err)

	Unpublish("1")
}

func TestPublish(t *testing.T) {
	created := Release{
		Version:     "0.0.1",
		Description: "Test publish",
	}
	buf := bytes.NewBufferString("RELEASE")
	r, err := Publish(created, buf)

	assert.NoError(t, err)
	assert.NotNil(t, r)

	f, err := os.Open(dataFilePath(r.ID + ".json"))
	assert.NoError(t, err)
	fromJSON := Release{}
	err = json.NewDecoder(f).Decode(&fromJSON)
	assert.NoError(t, err)
	assert.Equal(t, created.Version, fromJSON.Version)
	assert.Equal(t, created.Description, fromJSON.Description)

	fData, err := os.Open(dataFilePath(r.ID + ".dat"))
	assert.NoError(t, err)
	c, err := ioutil.ReadAll(fData)
	assert.NoError(t, err)
	assert.Equal(t, "RELEASE", string(c))

	f.Close()
	fData.Close()

	assert.NoError(t, Unpublish(r.ID))
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

	_, err = os.Stat(dataFilePath(r.ID + ".json"))
	assert.True(t, os.IsNotExist(err))
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
