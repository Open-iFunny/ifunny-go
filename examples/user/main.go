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
	u, err := client.GetUser(compose.UserByNick("gastrodon"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("[%s (%s)]: %s\n", u.Nick, u.ID, u.About)
}
