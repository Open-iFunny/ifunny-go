package compose

import "github.com/gastrodon/turnpike"

/*
call out to get chat contacts
*/
func Contacts(limit int) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("list_contacts"),
		ArgumentsKw: map[string]interface{}{"limit": limit},
	}
}

func SearchContacts(query string, limit int) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("search_contacts"),
		ArgumentsKw: map[string]interface{}{"query": query, "limit": limit},
	}
}

func Operators(channel string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("list_operators"),
		ArgumentsKw: map[string]interface{}{"chat_name": channel},
	}
}

func UserByID(id string) Request {
	return get("/users/"+id, nil)
}

func UserByNick(nick string) Request {
	return get("/users/by_nick/"+nick, nil)
}

func UserAccount() Request {
	return get("/account", nil)
}
