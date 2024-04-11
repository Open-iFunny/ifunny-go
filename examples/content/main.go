package main

import (
	"fmt"
	"os"

	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
)

var bearer = os.Getenv("IFUNNY_BEARER")
var userAgent = os.Getenv("IFUNNY_USER_AGENT")

func printContent(c *ifunny.Content) {
	fmt.Printf("[%s @ %d]: tags: %v, smiles: %d\n",
		c.Creator.Nick, c.PublushAt, c.Tags, c.Num.Smiles)
}

func main() {
	client, _ := ifunny.MakeClient(bearer, userAgent)
	content, err := client.GetContent("lgzM46Im9")
	if err != nil {
		panic(err)
	}

	printContent(content)

	feed := "featured"
	page, err := client.GetFeedPage(compose.Feed(feed, 5, compose.NoPage[string]()))
	if err != nil {
		panic(err)
	}

	fmt.Printf("got %d items from the feed!\n", len(page.Items))
	for _, c := range page.Items {
		printContent(&c)
	}

	if !page.Paging.HasNext {
		fmt.Println("that was the end of the feed")
	}

	page, err = client.GetFeedPage(compose.Feed(feed, 5, compose.Next(page.Paging.Cursors.Next)))
	if err != nil {
		panic(err)
	}

	fmt.Printf("got %d MORE items from the feed!\n", len(page.Items))
	for _, c := range page.Items {
		printContent(&c)
	}
}
