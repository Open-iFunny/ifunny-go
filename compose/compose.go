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

type Page[T int | string] struct {
	Key   pageDirection
	Value T
}

func NoPage[T int | string]() Page[T]      { return Page[T]{NONE, *new(T)} }
func Prev[T int | string](value T) Page[T] { return Page[T]{PREV, value} }
func Next[T int | string](value T) Page[T] { return Page[T]{NEXT, value} }

func get(path string, query url.Values) Request {
	return Request{Method: "GET", Path: path, Query: query}
}
