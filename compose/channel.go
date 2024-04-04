package compose

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gastrodon/turnpike"
)

type ChannelType int

const (
	ChannelDM      ChannelType = 1
	ChannelPrivate ChannelType = 2
	ChannelPublic  ChannelType = 3
)

type ChannelJoinState int

const (
	NotJoined ChannelJoinState = 0
	Invited   ChannelJoinState = 1
	Joined    ChannelJoinState = 2
)

type ChannelRole int

const (
	RoleDM     ChannelRole = 0 // ?
	RoleNormie ChannelRole = 2 // ???
)

func JoinedChannels(id string) turnpike.Subscribe {
	return turnpike.Subscribe{Topic: URI("user." + id + ".chats")}
}

func HideChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("hide_chat"),
		ArgumentsKw: map[string]interface{}{"chat_name": channel},
	}
}

func CreateChannel(kind ChannelType, id, title, description, coverURL string, invitedIDs []string) turnpike.Call {
	call := turnpike.Call{
		Procedure: URI("create_channel"),
		ArgumentsKw: map[string]interface{}{
			"type":             kind,
			"id":               id,
			"title":            title,
			"description":      description,
			"coverURL":         coverURL,
			"inviteMembersIDs": invitedIDs,
		},
	}

	return call
}

func DMChannelName(self string, them []string) string {
	us := append(them, self)
	sort.Strings(us)
	size := len(us)
	backwards := make([]string, size)
	for index, each := range us {
		backwards[size-1-index] = each
	}

	return strings.Join(backwards, "_")
}

func GetDMChannel(id string, them ...string) turnpike.Call {
	return turnpike.Call{
		Procedure: URI("get_or_create_chat"),
		ArgumentsKw: map[string]interface{}{
			"type":  ChannelDM,
			"users": them,
			"name":  DMChannelName(id, them),
		},
	}
}

func NewChannel(title, name, description string, invite []string, channelType ChannelType) turnpike.Call {
	if description != "" && channelType == ChannelPrivate {
		panic("cannot add a description to a private channel")
	}

	return turnpike.Call{
		Procedure: URI("new_chat"),
		ArgumentsKw: map[string]interface{}{
			"users":       invite,
			"title":       title,
			"name":        name,
			"description": description,
			"type":        channelType,
		},
	}
}

func GetChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("get_chat"),
		ArgumentsKw: map[string]interface{}{"chat_name": channel},
	}
}

func JoinChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("join_chat"),
		ArgumentsKw: map[string]interface{}{"chat_name": channel},
	}
}

func ExitChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("leave_chat"),
		ArgumentsKw: map[string]interface{}{"chat_name": channel},
	}
}

var (
	ChatsTrending = Request{Method: "GET", Path: "/chats/trending"}
)

func ChatsQuery(query string, limit int, page Page[string]) Request {
	return Request{
		Method: "GET", Path: "/chats/open_channels",
		Query: url.Values{
			"q":              []string{query},
			"limit":          []string{strconv.Itoa(limit)},
			string(page.Key): []string{page.Value},
		},
	}
}
