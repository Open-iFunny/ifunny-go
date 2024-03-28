package compose

import "github.com/gastrodon/turnpike"

func inviteCall(accept bool) string {
	if accept {
		return "invite.accept"
	}

	return "invite.decline"
}

func Invite(channel string, users []string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("invite.invite"),
		ArgumentsKw: map[string]interface{}{"chat_name": channel, "users": users},
	}
}

func InviteResponse(channel string, accept bool) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI(inviteCall(accept)),
		ArgumentsKw: map[string]interface{}{"chat_name": channel},
	}
}

func ReceiveInvite(id string) turnpike.Subscribe {
	return turnpike.Subscribe{Topic: URI("user." + id + ".invites")}
}

func Kick(channel, user string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("kick_member"),
		ArgumentsKw: map[string]interface{}{"user_id": user, "chat_name": channel},
	}
}
