package dist

import (
	"log"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

var dataFile = "./data/sc2a.db"
var db *bolt.DB

func mustWriteDefaultConfig(b *bolt.Bucket, name, defaultValue string) {
	v := b.Get([]byte(name))
	if v == nil {
		err := b.Put([]byte(name), []byte(defaultValue))
		if err != nil {
			log.Fatal(err)
		}
	}
}

//OpenDB opens database
func OpenDB() {
	var err error
	if err = os.MkdirAll(filepath.Dir(dataFile), 0600); err != nil {
		log.Fatal(err)
	}

	db, err = bolt.Open(dataFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	buckets := []string{"release", "sub", "link", "sub_download", "config"}
	db.Update(func(tx *bolt.Tx) error {
		for _, b := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

//CloseDB closes database
func CloseDB() {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}
