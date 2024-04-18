package main

import (
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
	client, _ := ifunny.MakeClient(bearer, userAgent)
	comments, err := client.GetCommentPage(compose.Comments("r4mB8i4NA", 30, compose.NoPage[string]()))
	if err != nil {
		panic(err)
	}

	for _, c := range comments.Items {
		printComment(&c)
	}

	featured, err := client.GetFeedPage(compose.Feed("featured", 1, compose.NoPage[string]()))
	if err != nil {
		panic(err)
	}

	data := client.IterComments(featured.Items[0].ID)
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
