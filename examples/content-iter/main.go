package main

import (
	"fmt"
	"os"

	"github.com/open-ifunny/ifunny-go"
)

var bearer = os.Getenv("IFUNNY_BEARER")
var userAgent = os.Getenv("IFUNNY_USER_AGENT")

func printContent(c *ifunny.Content) {
	fmt.Printf("[%s @ %d]: tags: %v, smiles: %d\n",
		c.Creator.Nick, c.PublushAt, c.Tags, c.Num.Smiles)
}

func main() {
	client, _ := ifunny.MakeClient(bearer, userAgent)

	fmt.Println("iterating features")
	iter := client.IterFeed("featured")
	for i := 0; i < 60; i++ {
		r := <-iter
		if r.Err != nil {
			panic(r.Err)
		}

		if r.V == nil {
			break
		}

		printContent(r.V)
	}

	fmt.Println("iterating our timeline")
	iter = client.IterTimeline(client.Self.ID)
	for i := 0; i < 60; i++ {
		r := <-iter
		if r.Err != nil {
			panic(r.Err)
		}

		if r.V == nil {
			fmt.Println("broke")
			break
		}

		printContent(r.V)
	}
}
