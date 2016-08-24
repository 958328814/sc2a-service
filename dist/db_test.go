package dist

import "testing"
import "os"
import "log"

func TestMain(m *testing.M) {
	if err := os.RemoveAll(dataFile); err != nil {
		log.Fatal(err)
	}
	OpenDB()
	rv := m.Run()
	CloseDB()
	os.Exit(rv)
}
