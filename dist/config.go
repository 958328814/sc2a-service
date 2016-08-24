package dist

import (
	"encoding/json"

	"html/template"

	"github.com/boltdb/bolt"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

var mg mailgun.Mailgun
var makeLink func(string) string
var nameTemplate *template.Template
var notifyEmailSubjectTemplate *template.Template
var notifyEmailContentTemplate *template.Template

// Config contains some basic config options
type Config struct {
	FilenameTemplate           string
	NotifyEmailSubjectTemplate string
	NotifyEmailContentTemplate string
	Mailgun                    struct {
		Domain string
		APIKey string
	}
}

// Configure initializes this package
func Configure(baseURI string, c Config) {
	mg = mailgun.NewMailgun(c.Mailgun.Domain, c.Mailgun.APIKey, "")
	makeLink = func(id string) string {
		return baseURI + "/download/" + id
	}
	nameTemplate = template.Must(template.New("nameTemplate").Parse(c.FilenameTemplate))
	notifyEmailSubjectTemplate = template.Must(template.New("notifyEmailSubjectTemplate").Parse(c.NotifyEmailSubjectTemplate))
	notifyEmailContentTemplate = template.Must(template.New("notifyEmailContentTemplate").Parse(c.NotifyEmailContentTemplate))
}

// ConfigMap is a alias of map[string]interface{}
type ConfigMap map[string]interface{}

// GetAllConfig loads all config values from db
func GetAllConfig() (ConfigMap, error) {
	rv := ConfigMap{}
	err := db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("config")).ForEach(func(k, v []byte) error {
			var vv interface{}
			err := json.Unmarshal(v, &vv)
			if err != nil {
				return err
			}
			rv[string(k)] = vv
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return rv, nil
}

// UpdateConfig merges submit conifg values with db
func UpdateConfig(values ConfigMap) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		for k, v := range values {
			j, err := json.Marshal(v)
			if err != nil {
				return err
			}
			err = b.Put([]byte(k), j)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
