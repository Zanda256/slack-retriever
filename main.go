package main

import (
	"github.com/Zanda256/slack-retriever/slack"
)

func main() {
	allGood := slack.ParseArgs()
	if !allGood {
		fmt.Println("cannot parse arguments, ChannelID and API token are both required.")
	}
	//Initialize rate limited http client
	c := slack.NewClient()
	//Initialize elasticsearch Client
	es7 := slack.NewESClient()

	myFetcher := slack.NewFetcher("slack", "1.0", c, es7, Archives)

	channelInfo, err := f.GetChannelInfo()
	checkError(err)

	fmt.Println(channelInfo)

	if channelInfo["arch"] == true && Archives == false {
		fmt.Println("Channel %s is archived. Set Archives option to true to fetch messages", slack.ChID)
	}

	p := &msgHistParams{}

	numChannelMembers, err := GetChannelMembers(myFetcher, p)
	checkError(err)
	fmt.Println(numChannelMembers)
	channelInfo["numMembers"] = numChannelMembers

	msgs := make([]slack.RawMsg)
	msgs = GetMsgHistory(myFetcher, p)

}

func checkError(err) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
