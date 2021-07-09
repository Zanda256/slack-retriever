package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
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
	ElasticClient   *elasticsearch.Client
	BackendVersion  string
	Debug           int
}

//NewFetcher creates a Fetcher instance
func NewFetcher(dsName, backEndVers string, webClient *Client, esClient *elasticsearch.Client, allowArchive bool, debug int) *Fetcher {
	f := &Fetcher{
		DSName:          dsName,
		BackendVersion:  backEndVers,
		HTTPClient:      webClient,
		ElasticClient:   esClient,
		IncludeArchived: allowArchive,
		Debug:           debug,
	}
	return f
}

//GetChannelInfo method makes the conversations.info api call
func (f *Fetcher) GetChannelInfo() (map[string]interface{}, error) {
	var rData convoInfoRawResponse
	Info := make(map[string]interface{})
	par := &MsgHistParams{
		ChannelID: ChID,
		Token:     APIToken,
		EndPoint:  convoInfo,
	}
	fmt.Printf("par.-%s , CHID : %s\n", par.ChannelID, ChID)
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
		fmt.Printf("Result : %+v", rData)
		err = fmt.Errorf("can not fetch channel info because : %s", rData.Err)
		return nil, err
	}
	return Info, nil
}

//GetChannelMembers fetches a list of memberIds for given channel
func GetChannelMembers(f *Fetcher, par *MsgHistParams) (int, error) {
	var (
		rData      convoMembersRawResponse
		numMembers int
	)
	fetchMembers := func() (*http.Response, error) {
		par.ChannelID = ChID
		par.Token = APIToken
		par.EndPoint = convoMembers

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
		par.Cursor = ok
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
func GetMsgHistory(f *Fetcher, par *MsgHistParams) ([]RawMsg, error) {
	var (
		rData    convoHistoryResponse
		Messages []RawMsg
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

		par.ChannelID = ChID
		par.Token = APIToken
		par.EndPoint = convoMessages
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
		if MaxItems < 0 {
			break
		}
		par.Cursor = ok
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

func (f *Fetcher) makeAPICall(params *MsgHistParams) (*http.Response, error) {
	endPointURL := strings.Join([]string{slackAPIURL, params.EndPoint}, "/")
	tokenstr := fmt.Sprintf("Bearer %s", params.Token)
	usrAgentstr := fmt.Sprintf("Go_Perceaval_%s/%s", f.DSName, f.BackendVersion)

	req, err := http.NewRequest("GET", endPointURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//setting request headers
	req.Header.Set("User-Agent", usrAgentstr)
	req.Header.Set("Authorization", tokenstr)
	req.Header.Set("Content-type", "application/json; charset=utf-8")

	//setting request query string parameters
	q := req.URL.Query()
	q.Add("channel", params.ChannelID)
	if params.EndPoint == convoMessages {
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
		if params.Cursor != "" {
			q.Add("cursor", params.Cursor)
		}
		if MaxItems > 1000 {
			fmt.Println("Maximum number of items that can be retrieved is 1000.")
			MaxItems = 1000
			params.limit = 200
			MaxItems -= 200
		} else if 1000 >= MaxItems && MaxItems >= 200 {
			params.limit = 200
			MaxItems -= 200
		} else if 0 < MaxItems && MaxItems < 200 {
			params.limit = int(MaxItems)
			MaxItems -= 200
		}
		lmt := strconv.Itoa(params.limit)
		q.Add("limit", lmt)
	}
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

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
