package slack

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

// "service_name": "Leg end nary a laugh, Ink.",
// "text": "This is likely a pun about the weather.",
// "fallback": "We're withholding a pun from you",
// "thumb_url": "https://badpuns.example.com/puns/123.png",
// "thumb_width": 1920,
// "thumb_height": 700,
// "id": 1

var RawMapping = `{"mappings":
					{"dynamic":true, 
						"properties": 
						{   "ok":{"type":"bool"}, 
							"messages": {"type": "nested", 
								"properties" : {
									"type":{"type":"keyword"}
									"user":{"type":"keyword"}
									"text":{"type":"text"}
									"ts":{"type":"float"}
									"attachments":{"type":"nested",
										"properties": {
											"service_name":{"type":"text"}
											"text": {"type":"text"}
											"fallback": {"type":"text"}
											"thumb_url": {"type":"text"}
											"thumb_width": {"type":"integer"}
											"thumb_height": {"type":"integer"}
											"id": {"type":"integer"}
										}
									}
								}
							}
						}
					}
				}`

//NewESClient ctreates anew elasticsearch client
func NewESClient() (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Printf("Error creating the client: %s", err)
		return nil, err
	}
	return es, nil
}
