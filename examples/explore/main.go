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

func printUser(u *ifunny.User) {
	fmt.Printf("[%s (%s)]: %d ->, %d <-\n", u.Nick, u.ID, u.Num.Subscribers, u.Num.Subscriptions)
}

func main() {
	client, _ := ifunny.MakeClient(bearer, userAgent)

	page, err := client.ExploreContentPage(compose.Explore("content_shuffle", 30, compose.NoPage[string]()))
	if err != nil {
		panic(err)
	}

	fmt.Printf("got %d content items from explore!\n", len(page.Items))
	for _, c := range page.Items {
		printContent(&c)
	}

	contentIter := client.IterExploreContent("category-science-tech")
	for i := 0; i < 120; i++ {
		r := <-contentIter
		if r.Err != nil {
			panic(r.Err)
		}

		if r.V == nil {
			fmt.Println("broke")
			break
		}

		printContent(r.V)
	}

	userIter := client.IterExploreUser("users_top_by_subscribers")
	for i := 0; i < 60; i++ {
		r := <-userIter
		if r.Err != nil {
			panic(r.Err)
		}

		if r.V == nil {
			fmt.Println("broke")
			break
		}

		printUser(r.V)
	}
}
