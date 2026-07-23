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

func printComment(c *ifunny.Comment) {
	fmt.Printf("[%s @ %d]: %s, smiles: %d\n",
		c.User.Nick, c.Date, c.Text, c.Num.Smiles)
}

func main() {
	ctx := context.Background()
	client, _ := ifunny.MakeClient(ctx, bearer, ifunny.RawUserAgent(userAgent))
	comments, err := ifunny.FetchPage[ifunny.CommentsEnvelope](ctx, client, compose.Comments("r4mB8i4NA").Request(compose.NoPage()))
	if err != nil {
		panic(err)
	}

	for _, c := range comments.Items {
		printComment(&c)
	}

	featuredFeed := compose.NamedFeed("featured")
	featuredFeed.Limit = 1
	featured, err := ifunny.FetchPage[ifunny.FeedEnvelope](ctx, client, featuredFeed.Request(compose.NoPage()))
	if err != nil {
		panic(err)
	}

	data := client.IterCommentsRoots(ctx, featured.Items[0].ID)
	for i := 0; i < 60; i++ {
		r := <-data
		if r.Err != nil {
			panic(r.Err)
		}

		if r.V == nil {
			break
		}

		printComment(r.V)
	}
}
