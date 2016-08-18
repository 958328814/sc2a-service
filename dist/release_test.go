package dist

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	assert.NoError(t, os.RemoveAll(DataDir))
	assert.NoError(t, os.MkdirAll(DataDir, 0700))
	assert.NoError(t, ioutil.WriteFile(DataDir+"/1.json", []byte("1"), 0755))
	assert.NoError(t, ioutil.WriteFile(DataDir+"/1.dat", []byte("1"), 0755))

	items, err := List()
	assert.NoError(t, err)
	log.Println(items)
}
