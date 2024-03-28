package compose

import (
	"io"
	"net/url"

	"github.com/gastrodon/turnpike"
)

const chatNamespace = "co.fun.chat"

func URI(name string) turnpike.URI { return turnpike.URI(chatNamespace + "." + name) }

type Request struct {
	Method, Path string
	Body         io.Reader
	Query        url.Values
}

type pageDirection string

const (
	NONE pageDirection = ""
	NEXT pageDirection = "next"
	PREV pageDirection = "prev"
)

type Page struct {
	Key   pageDirection
	Value int64
}

func NoPage() Page          { return Page{NONE, 0} }
func Prev(value int64) Page { return Page{PREV, value} }
func Next(value int64) Page { return Page{NEXT, value} }

func get(path string, query url.Values) Request {
	return Request{Method: "GET", Path: path, Query: query}
}

type SPage struct {
	Key   pageDirection
	Value string
}
