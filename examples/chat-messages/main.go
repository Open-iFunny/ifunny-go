package main

import (
	"fmt"
	"os"

	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
)

var bearer = os.Getenv("IFUNNY_BEARER")
var userAgent = os.Getenv("IFUNNY_USER_AGENT")

func main() {
	client, _ := ifunny.MakeClient(bearer, userAgent)
	chat, _ := client.Chat()

	channels, err := client.GetChannels(compose.ChatsTrending)
	if err != nil {
		panic(err)
	}

	fmt.Printf("got %d trendy chat channels!\n", len(channels))

	messages, _, _, err := chat.ListMessages(compose.ListMessages("apitools", 10, compose.NoPage[int]()))
	if err != nil {
		panic(err)
	}

	fmt.Printf("got %d messages from apitools...\n", len(messages))
	for _, m := range messages {
		fmt.Printf("[%.0f] %s: %s\n", m.PubAt, m.User.Nick, m.Text)
	}

	messages, _, _, err = chat.ListMessages(compose.ListMessages("apitools", 10, compose.Prev(0)))
	if err != nil {
		panic(err)
	}

	fmt.Printf("got %d oldest 10 messages from apitools...\n", len(messages))
	for _, m := range messages {
		fmt.Printf("[%.0f] %s: %s\n", m.PubAt, m.User.Nick, m.Text)
	}

}
