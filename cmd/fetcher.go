package slack

import (
	"fmt"
	"net/http"
	"strconv"
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
func (f *Fetcher) GetChannelInfo() (*http.Response, error) {
	par := &msgHistParams{
		channelID: ChID,
		token:     APIToken,
		endPoint:  convoInfo,
	}
	r, err := f.makeAPICall(par)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return r, nil
}

//GetChannelMembers fetches a list of memberIds for given channel
func (f *Fetcher) GetChannelMembers(par *msgHistParams) (*http.Response, error) {
	par.channelID = ChID
	par.token = APIToken
	par.endPoint = convoMessages
	r, err := f.makeAPICall(par)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return r, nil
}

//GetMsgHistory fetches messages from the specified channel
func (f *Fetcher) GetMsgHistory(par *msgHistParams) (*http.Response, error) {
	if Oldest != "" {
		frm, err := dateStrToUnix(Oldest)
		if err != nil {
			return nil, err
		}
		par.oldest = frm
	}
	if Latest != "" {
		to, err := dateStrToUnix(Latest)
		if err != nil {
			return nil, err
		}
		par.latest = to
	}

	par.channelID = ChID
	par.token = APIToken
	par.endPoint = convoMessages
	r, err := f.makeAPICall(par)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (f *Fetcher) makeAPICall(params *msgHistParams) (*http.Response, error) {
	endPointURL := strings.Join([]string{slackAPIURL, params.endPoint}, "/")
	tokenstr := fmt.Sprintf("Bearer %s", params.token)
	usrAgentstr := fmt.Sprintf("Go_Perceaval_%s/%s", f.DSName, f.BackendVersion)

	req, err := http.NewRequest("GET", endPointURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//setting request headers
	req.Header.Set("User-Agent", usrAgentstr)
	req.Header.Set("Authorization", tokenstr)

	//setting request query string parameters
	q := req.URL.Query()
	q.Add("channel", params.channelID)
	if params.endPoint == convoMessages {
		if params.oldest > 0 {
			fromDt := strconv.FormatFloat(params.oldest, 'f', 6, 64)
			q.Add("oldest", fromDt)
			q.Add("inclusive", "true")
		}
		if params.latest > 0 {
			ToDt := strconv.FormatFloat(params.latest, 'f', 6, 64)
			q.Add("latest", ToDt)
			q.Add("inclusive", "true")
		}
		if params.cursor != "" {
			q.Add("cursor", params.cursor)
		}
		lmt := strconv.Itoa(params.limit)
		q.Add("limit", lmt)
	}
	if params.cursor != "" {
		q.Add("cursor", params.cursor)
	}

	resp, err := f.HTTPClient.DoRequest(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return resp, nil
}

func dateStrToUnix(dateFromCLI string) (float64, error) {
	lst := strings.Split(dateFromCLI, "-")
	if len(lst) < 3 {
		err := fmt.Errorf("could not parse date %s to unix time", dateFromCLI)
		return -1, err
	}
	dateStr := fmt.Sprintf("%s-%s-%s", lst[2], lst[1], lst[0])
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	tUnix := float64(t.Unix())
	return tUnix, nil
}
