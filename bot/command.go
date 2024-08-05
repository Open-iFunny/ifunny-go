package bot

import (
	"strings"
)

type prefix struct {
	prefix string
}

func (us filter) And(also filter) filter {
	return func(ctx Context) (bool, error) {
		if ok, err := us(ctx); ok && err == nil {
			return also(ctx)
		} else {
			return ok, err
		}
	}
}

func (us filter) Not(also filter) filter {
	return func(ctx Context) (bool, error) {
		if ok, err := us(ctx); ok && err == nil {
			ok, err = also(ctx)
			return !ok, err
		} else {
			return ok, err
		}
	}
}

func Prefix(fix string) prefix { return prefix{fix} }

func (fix prefix) Cmd(name string) filter {
	return func(ctx Context) (bool, error) {
		event, err := ctx.Event()
		if err != nil {
			return false, err
		}

		return event.Text == fix.prefix+name || strings.HasPrefix(event.Text, fix.prefix+name+" "), nil
	}
}

func (fix prefix) Any() filter {
	return func(ctx Context) (bool, error) {
		event, err := ctx.Event()
		if err != nil {
			return false, err
		}

		return strings.HasPrefix(event.Text, fix.prefix), nil
	}
}

func AuthoredBy(id string) filter {
	return func(ctx Context) (bool, error) {
		if caller, err := ctx.Caller(); err != nil {
			return false, err
		} else {
			return caller.ID == id, nil
		}
	}
}
