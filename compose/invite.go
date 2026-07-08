package compose

import "github.com/gastrodon/turnpike"

// Invite composes a call to invite users to a channel.
func Invite(channel string, users []string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("invite.invite"),
		ArgumentsKw: map[string]any{"chat_name": channel, "users": users},
	}
}

// InviteResponse composes a call to accept or decline a channel invitation.
func InviteResponse(channel string, accept bool) turnpike.Call {
	proc := "invite.decline"
	if accept {
		proc = "invite.accept"
	}
	return turnpike.Call{
		Procedure:   URI(proc),
		ArgumentsKw: map[string]any{"chat_name": channel},
	}
}

// ReceiveInvite subscribes to channel invitations for a user.
func ReceiveInvite(id string) turnpike.Subscribe {
	return turnpike.Subscribe{Topic: URI("user." + id + ".invites")}
}

// Kick composes a call to remove a user from a channel.
func Kick(channel, user string) turnpike.Call {
	return turnpike.Call{
		Procedure:   URI("kick_member"),
		ArgumentsKw: map[string]any{"user_id": user, "chat_name": channel},
	}
}
