package slack

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var RawMapping = `{"mappings":
					{"dynamic":true, 
						"properties": 
						{   "ok":{"type":"bool"}, 
							"messages": {"type": "nested", 
								"properties" : {
									"type":{"type":"keyword"},
									"user":{"type":"keyword"},
									"text":{"type":"text"},
									"ts":{"type":"float"},
									"attachments":{"type":"nested",
										"properties": {
											"service_name":{"type":"text"},
											"text": {"type":"text"},
											"fallback": {"type":"text"},
											"thumb_url": {"type":"text"},
											"thumb_width": {"type":"integer"},
											"thumb_height": {"type":"integer"},
											"id": {"type":"integer"}
										}
									}
								}
							}
						}
					}
				}`

// EsStorage has es client as a field
type EsStorage struct {
	c *elasticsearch.Client
}

//NewEsStorage ctreates anew elasticsearch client
func NewEsStorage() (*EsStorage, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return nil, err
	}
	s := &EsStorage{es}
	return s, nil
}

//NewESIndex creates a new Index called "name" with mappings as "mapping"
func (s *EsStorage) NewESIndex(name, mapping string) (bool, error) {
	b := strings.NewReader(mapping)
	r := &esapi.IndicesCreateRequest{
		Index: name,
		Body:  b,
	}
	ctx := context.Background()
	resp, err := r.Do(ctx, s.c)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if resp.IsError() {
		err = fmt.Errorf("Failed to create index: %d %s", resp.StatusCode, name)
		return false, err
	}
	fmt.Println(resp.String())
	return true, nil
}
