package compose

import (
	"io"
	"net/http"
	"net/url"

	"github.com/gastrodon/turnpike"
)

const chatNamespace = "co.fun.chat"

func URI(name string) turnpike.URI { return turnpike.URI(chatNamespace + "." + name) }

type Request struct {
	Method, Path string
	Body         io.Reader
	Query        url.Values
	Header       http.Header
}

type pageDirection string

const (
	NONE pageDirection = ""
	NEXT pageDirection = "next"
	PREV pageDirection = "prev"
)

// Value is a page token value. Its String method produces the wire form sent to
// the API (a raw cursor for [Literal], or an encoded exclusion set for [IDs]).
type Value interface{ String() string }

// Literal wraps a plain string page token — the historical behavior, and what
// non-collective feeds use. Its String is just the underlying value. (The WAMP
// list_messages cursor is numeric and does not go through Value; see
// [ListMessages].)
type Literal struct{ Wrapped string }

func (l Literal) String() string { return l.Wrapped }

// Page is one pagination step: a direction (Key) and the token Value to send.
// The zero Page (NoPage) requests the first page.
type Page struct {
	Key   pageDirection
	Value Value
}

func NoPage() Page          { return Page{NONE, nil} }
func Prev(value Value) Page { return Page{PREV, value} }
func Next(value Value) Page { return Page{NEXT, value} }

func get(path string, query url.Values) Request {
	return Request{Method: "GET", Path: path, Query: query}
}
