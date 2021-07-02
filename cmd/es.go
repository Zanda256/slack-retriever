package slack

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func newESClient() (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return nil, err
	}
	return es, nil
}
