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

// JoinedChannels subscribes to updates for a user's joined channels.
func JoinedChannels(id string) turnpike.Subscribe {
	return turnpike.Subscribe{Topic: URI("user." + id + ".chats")}
}

// HideChannel composes a call to hide a channel from the user's view.
func HideChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("hide_chat"),
		ArgumentsKw: map[string]any{"chat_name": channel},
	}
}

// CreateChannel composes a call to create a new channel with the specified type and members.
func CreateChannel(kind ChannelType, id, title, description, coverURL string, invitedIDs []string) turnpike.Call {
	call := turnpike.Call{
		Procedure: URI("create_channel"),
		ArgumentsKw: map[string]any{
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

// DMChannelName generates a canonical DM channel name from participant user IDs.
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

// GetDMChannel composes a call to get or create a direct message channel with specified users.
func GetDMChannel(id string, them ...string) turnpike.Call {
	return turnpike.Call{
		Procedure: URI("get_or_create_chat"),
		ArgumentsKw: map[string]any{
			"type":  ChannelDM,
			"users": them,
			"name":  DMChannelName(id, them),
		},
	}
}

// NewChannel composes a call to create a new chat channel with the specified type and members.
func NewChannel(title, name, description string, invite []string, channelType ChannelType) turnpike.Call {
	if description != "" && channelType == ChannelPrivate {
		panic("cannot add a description to a private channel")
	}

	return turnpike.Call{
		Procedure: URI("new_chat"),
		ArgumentsKw: map[string]any{
			"users":       invite,
			"title":       title,
			"name":        name,
			"description": description,
			"type":        channelType,
		},
	}
}

// GetChannel composes a call to get channel details by name.
func GetChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("get_chat"),
		ArgumentsKw: map[string]any{"chat_name": channel},
	}
}

// JoinChannel composes a call to join a channel by name.
func JoinChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("join_chat"),
		ArgumentsKw: map[string]any{"chat_name": channel},
	}
}

// ExitChannel composes a call to leave a channel by name.
func ExitChannel(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("leave_chat"),
		ArgumentsKw: map[string]any{"chat_name": channel},
	}
}

var (
	ChatsTrending = Request{Method: "GET", Path: "/chats/trending"}
)

// ChatsQuery composes a request to search for open channels by query string with pagination.
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
