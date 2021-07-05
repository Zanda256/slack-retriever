package slack

import (
	"flag"
)

//ParseArgs function to parse command line arguments
func ParseArgs() bool {
	flag.StringVar(&ChID, "ChannelID", "", "ID of the channel to fetch messages from. Usage: -ChannelID C1234567890")
	flag.StringVar(&APIToken, "Token", "", "slack API token. Usage: -Token xxxx-xxxxxxxxx-xxxx")
	flag.StringVar(&Oldest, "Oldest", "", "Date from which to retrive messages in the format: -Latest 02-01-2006 meaning Day(of the month)/Month/Year")
	flag.StringVar(&Latest, "Latest", "", "Date up to which to retrive messages in the format: -Latest 02-01-2006 meaning Day(of the month)/Month/Year")
	flag.Uint64Var(&MaxItems, "MaxItems", 1000, "Maximum number of messages to retrive. Usage : -MaxItems 200")
	flag.BoolVar(&Archives, "Archives", false, "Set this option to true to retreive messages from an archived channel. Usage : -Archives true")
	flag.Parse()
	if ChID == "" {
		return false
	} else if APIToken == "" {
		return false
	}
	return flag.Parsed()
}
