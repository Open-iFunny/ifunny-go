# ifunny-go

Go bindings for reading the iFunny HTTP + chat APIs. API reference on pkg.go.dev: **[`ifunny-go`](https://pkg.go.dev/github.com/open-ifunny/ifunny-go)** · [`compose`](https://pkg.go.dev/github.com/open-ifunny/ifunny-go/compose) · [`bot`](https://pkg.go.dev/github.com/open-ifunny/ifunny-go/bot)

## Install

```
go get github.com/open-ifunny/ifunny-go
```

## Quickstart

Anonymous, read-only client via a minted basic token:

```go
package main

import (
	"context"
	"fmt"

	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
)

func main() {
	ctx := context.Background()
	basic, _ := ifunny.GenerateBasic()
	ua := ifunny.Android{Version: "14"}.UserAgent()

	client, _ := ifunny.MakeClientBasic(basic, ua)
	if err := client.PrimeBasic(ctx); err != nil { // ~15s server-side wait
		panic(err)
	}

	u, err := client.GetUser(ctx, compose.UserByNick("woof"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s (%s)\n", u.Nick, u.ID)
}
```

For an authenticated client (chat, personalized feeds, `Self`), pass a bearer token to `ifunny.MakeClient` instead. Every HTTP-backed accessor takes a `context.Context` as its first argument, so requests honor cancellation and deadlines; cancelling the ctx also stops an in-flight `Iter*` pager. Iterators over feeds, comments, users, and chat channels live as `Iter*` methods on `*Client` and `*Chat`. See [`examples/`](examples/) for more.

## TODO
- [x] Move over the chat library code from [discovery-bot](https://github.com/open-ifunny/discovery-bot)
- [x] Move over the feed scouring code from psyduck-etl/ifunny
- [x] Basic crawling: content (feeds, explore, timelines), users, comments
- [ ] Migrate iteration channels to Go 1.23+ range-over-func iterators + replace the current `chan Result[T]` streaming shape

## Thanks

- [@makeshiftartist](https://github.com/makeshiftartist) for the basic-auth research + impl in [`iFunny.ts`](https://github.com/makeshiftartist/ifunny.ts)
