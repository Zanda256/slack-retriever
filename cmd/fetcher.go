package slack

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
// 	slackAPIURL   = "https://slack.com/api"
// 	convoInfo     = "conversations.info"
// 	convoMembers  = "conversations.members"
// 	convoMessages = "conversations.history"
)

//Fetcher manges the fetching process
type Fetcher struct {
	DSName          string
	IncludeArchived bool
	HTTPClient      *Client
	ElasticClient   interface{}
	BackendVersion  string
	Debug           int
	DateFrom        time.Time
}

//GetChannelInfo method makes the conversations.info api call
func (f *Fetcher) GetChannelInfo(apiToken, channelID string) (*http.Response, error) {
	chanInfoURL := strings.Join([]string{slackAPIURL, convoInfo}, "/")
	tokenstr := fmt.Sprintf("Bearer %s", apiToken)
	req, err := http.NewRequest("GET", chanInfoURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("Authorization", tokenstr)
	q := req.URL.Query()
	q.Add("channel", channelID)
	resp, err := f.HTTPClient.DoRequest(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return resp, nil
}
