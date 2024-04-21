package main

import (
	"fmt"
	"os"

	"github.com/open-ifunny/ifunny-go"
)

var bearer = os.Getenv("IFUNNY_BEARER")
var userAgent = os.Getenv("IFUNNY_USER_AGENT")

func printChannel(c *ifunny.ChatChannel) {
	fmt.Printf("%s: %d/%d online\n",
		c.Title, c.MembersOnline, c.MembersTotal)
}

func main() {
	q := "hello"
	client, _ := ifunny.MakeClient(bearer, userAgent)

	fmt.Printf("iterating results for q=%s\n", q)
	iter := client.IterChannelsQuery(q)
	for i := 0; i < 60; i++ {
		r := <-iter
		if r.Err != nil {
			panic(r.Err)
		}

		if r.V == nil {
			fmt.Println("broke")
			break
		}

		printChannel(r.V)
	}
}
