package main

import (
	"fmt"
	"os"

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
	// es7, err := slack.NewESClient()
	// checkError(err)

	myFetcher := slack.NewFetcher("slack", "1.0", c, slack.Archives, 0)

	channelInfo, err := myFetcher.GetChannelInfo()
	checkError(err)

	fmt.Printf("channelInfo: %+v\n", channelInfo)

	if channelInfo["arch"] == true && slack.Archives == false {
		fmt.Printf("Channel %s is archived. Set Archives option to true to fetch messages", slack.ChID)
	}

	p := &slack.MsgHistParams{}

	numChannelMembers, err := slack.GetChannelMembers(myFetcher, p)
	checkError(err)
	fmt.Printf("Num Members: %d\n", numChannelMembers)
	channelInfo["numMembers"] = numChannelMembers

	msgs := make([]slack.RawMsg, 0)
	msgs, _ = slack.GetMsgHistory(myFetcher, p)
	fmt.Println(msgs)

}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
