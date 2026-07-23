package main

import (
	"context"
	"fmt"
	"os"

	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
)

var bearer = os.Getenv("IFUNNY_BEARER")
var userAgent = os.Getenv("IFUNNY_USER_AGENT")

func printContent(c *ifunny.Content) {
	fmt.Printf("[%s @ %d]: tags: %v, smiles: %d\n",
		c.Creator.Nick, c.PublishAt, c.Tags, c.Num.Smiles)
}

func main() {
	ctx := context.Background()
	client, _ := ifunny.MakeClient(ctx, bearer, ifunny.RawUserAgent(userAgent))

	fmt.Println("iterating features")
	iter := client.IterContent(ctx, compose.NamedFeed("featured"))
	for range 60 {
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
	iter = client.IterTimeline(ctx, client.Self.ID)
	for range 60 {
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
