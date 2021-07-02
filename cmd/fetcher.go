package slack

import (
	"encoding/json"
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
func (f *Fetcher) GetChannelInfo() (map[string]interface{}, error) {
	var rData convoInfoRawResponse
	var Info map[string]interface{}
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
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&rData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if rData.Success {
		Info["arch"] = rData.Channel["is_archived"].(bool)
		Info["Name"] = rData.Channel["name"].(string)
		Info["createdAt"] = rData.Channel["created"].(float64)
		Info["creator"] = rData.Channel["creator"].(string)

	} else if !rData.Success {
		err = fmt.Errorf("can not fetch channel info because : %s", rData.Err)
		return nil, err
	}
	return Info, nil
}

//GetChannelMembers fetches a list of memberIds for given channel
func GetChannelMembers(f *Fetcher, par *msgHistParams) (int, error) {
	var (
		rData      convoMembersRawResponse
		numMembers int
	)
	fetchMembers := func() (*http.Response, error) {
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
	resp, err := fetchMembers()
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	loadJSON := func(r *http.Response) error {
		err = json.NewDecoder(r.Body).Decode(&rData)
		defer r.Body.Close()
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}
	err = loadJSON(resp)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	if !rData.Success {
		err = fmt.Errorf("can not fetch channel members because : %s", rData.Err)
		return -1, err
	}
	numMembers += len(rData.MembersIDs)
	for ok := rData.ResponseMetadata.NextCursor; ok != ""; {
		par.cursor = ok
		resp, err := fetchMembers()
		if err != nil {
			fmt.Println(err)
			return -1, err
		}
		err = loadJSON(resp)
		if err != nil {
			fmt.Println(err)
			return -1, err
		}
		numMembers += len(rData.MembersIDs)
	}
	return numMembers, nil
}

//GetMsgHistory fetches messages from the specified channel
func GetMsgHistory(f *Fetcher, par *msgHistParams) ([]rawMsg, error) {
	var (
		rData    convoHistoryResponse
		Messages []rawMsg
	)
	fetchMsgs := func() (*http.Response, error) {
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
	resp, err := fetchMsgs()
	loadJSON := func(r *http.Response) error {
		err = json.NewDecoder(r.Body).Decode(&rData)
		defer r.Body.Close()
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}

	err = loadJSON(resp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	Messages = append(Messages, rData.Messages...)
	for ok := rData.ResponseMetadata.NextCursor; ok != ""; {
		par.cursor = ok
		resp, err := fetchMsgs()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		err = loadJSON(resp)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		Messages = append(Messages, rData.Messages...)
	}
	return Messages, nil
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
	}
	if params.endPoint == convoMessages || params.endPoint == convoMembers {
		if params.cursor != "" {
			q.Add("cursor", params.cursor)
		}
		if MaxItems > 1000 {
			fmt.Println("Maximum number of items that can be retrieved is 1000.")
			MaxItems = 1000
			params.limit = 200
			MaxItems -= 200
		} else if 1000 > MaxItems && MaxItems > 200 {
			params.limit = 200
			MaxItems -= 200
		} else if 0 < MaxItems && MaxItems < 200 {
			params.limit = int(MaxItems)
		}
		lmt := strconv.Itoa(params.limit)
		q.Add("limit", lmt)
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
