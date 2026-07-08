package compose

import "github.com/gastrodon/turnpike"

// Contacts composes a call to list chat contacts.
func Contacts(limit int) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("list_contacts"),
		ArgumentsKw: map[string]any{"limit": limit},
	}
}

// SearchContacts composes a call to search chat contacts by query.
func SearchContacts(query string, limit int) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("search_contacts"),
		ArgumentsKw: map[string]any{"query": query, "limit": limit},
	}
}

// Operators composes a call to list operators in a channel.
func Operators(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("list_operators"),
		ArgumentsKw: map[string]any{"chat_name": channel},
	}
}

// UserByID composes a request for user details by ID.
func UserByID(id string) Request {
	return get("/users/"+id, nil)
}

// UserByNick composes a request for user details by nick.
func UserByNick(nick string) Request {
	return get("/users/by_nick/"+nick, nil)
}

// UserAccount composes a request for the authenticated user's account details.
func UserAccount() Request {
	return get("/account", nil)
}

// Subscribers composes a request for a user's subscribers with pagination.
func Subscribers(id string, limit int, page Page[string]) Request {
	return get("/users/"+id+"/subscribers", feedParams(limit, page))
}

// Subscriptions composes a request for a user's subscriptions with pagination.
func Subscriptions(id string, limit int, page Page[string]) Request {
	return get("/users/"+id+"/subscriptions", feedParams(limit, page))
}
