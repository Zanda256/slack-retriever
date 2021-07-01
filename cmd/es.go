package slack

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

func newESClient() (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
}
