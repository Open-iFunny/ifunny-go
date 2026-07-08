# ifunny-go

Go bindings for reading the iFunny HTTP + chat APIs.

Full API reference on pkg.go.dev: **[`ifunny-go`](https://pkg.go.dev/github.com/open-ifunny/ifunny-go)** · [`compose`](https://pkg.go.dev/github.com/open-ifunny/ifunny-go/compose) · [`bot`](https://pkg.go.dev/github.com/open-ifunny/ifunny-go/bot)

## Install

```
go get github.com/open-ifunny/ifunny-go
```

## Quickstart

Anonymous, read-only client via a minted basic token:

```go
package main

import (
	"fmt"

	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
)

func main() {
	basic, _ := ifunny.GenerateBasic()
	ua := ifunny.Android{Version: "14"}.UserAgent()

	client, _ := ifunny.MakeClientBasic(basic, ua)
	if err := client.PrimeBasic(); err != nil { // ~15s server-side wait
		panic(err)
	}

	u, err := client.GetUser(compose.UserByNick("woof"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s (%s)\n", u.Nick, u.ID)
}
```

For an authenticated client (chat, personalized feeds, `Self`), pass a bearer token to `ifunny.MakeClient` instead. Iterators over feeds, comments, users, and chat channels live as `Iter*` methods on `*Client` and `*Chat`.

See [`examples/`](examples/) for more, including chat.

## Scope

Started as an auth-only helper and grew to cover the HTTP surface and the chat WAMP layer — this repo now supersedes [`open-ifunny/discovery-bot`](https://github.com/open-ifunny/discovery-bot).

## TODO
- [x] Move over the chat library code from [discovery-bot](https://github.com/open-ifunny/discovery-bot)
- [x] Move over the feed scouring code from psyduck-etl/ifunny
- [x] Basic crawling: content (feeds, explore, timelines), users, comments
- [ ] Migrate iteration channels to Go 1.23+ range-over-func iterators (Go now has first-class iterators; replace the current `chan Result[T]` streaming shape)

## Thanks

- [@makeshiftartist](https://github.com/makeshiftartist) for the basic-auth token algorithm in [ifunny.ts](https://github.com/makeshiftartist/ifunny.ts), ported into [`basic.go`](basic.go).
