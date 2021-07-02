package slack

var (
	ChID     string
	APIToken string
	Latest   string
	Oldest   string
	MaxItems uint64
	Archives bool
)

type msgHistParams struct {
	channelID string
	token     string
	cursor    string
	latest    float64
	limit     int
	oldest    float64
	endPoint  string
}

type convoInfoRawResponse struct {
	Success bool                   `json:"ok"`
	Channel map[string]interface{} `json:"channel"`
	Err     string                 `json:"error"`
}

type convoHistoryResponse struct {
	Success          bool     `json:"ok"`
	Messages         []rawMsg `json:"messages"`
	HasMore          bool     `json:"has_more"`
	PinCount         int      `json:"pin_count"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	Err string `json:"error"`
}

type convoMembersRawResponse struct {
	Success          bool     `json:"ok"`
	MembersIDs       []string `json:"members"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	Err string `json:"error"`
}

type msgAttachment struct {
	ServiceName string `json:"service_name"`
	Text        string `json:"text"`
	FallBack    string `json:"fallback"`
	ThumbURL    string `json:"thumb_url"`
	ThumbWidth  int    `json:"thumb_width"`
	ThumbHeight int    `json:"thumb_height"`
	ID          int    `json:"id"`
}

type rawMsg struct {
	Type        string          `json:"type"`
	UserID      string          `json:"user"`
	Text        string          `json:"text"`
	TimeStamp   float64         `json:"ts"`
	Attachments []msgAttachment `json:"attachments"`
}

// Messages []struct {
// 	Type        string  `json:"type"`
// 	UserID      string  `json:"user"`
// 	Text        string  `json:"text"`
// 	TimeStamp   float64 `json:"ts"`
// 	Attachments []struct {
// 		ServiceName string `json:"service_name"`
// 		Text        string `json:"text"`
// 		FallBack    string `json:"fallback"`
// 		ThumbURL    string `json:"thumb_url"`
// 		ThumbWidth  int    `json:"thumb_width"`
// 		ThumbHeight int    `json:"thumb_height"`
// 		ID          int    `json:"id"`
// 	} `json:"attachments"`
// } `json:"messages"`
