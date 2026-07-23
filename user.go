package ifunny

import (
	"context"

	"github.com/gastrodon/turnpike"
	"github.com/open-ifunny/ifunny-go/compose"
)

// User represents an iFunny user account. It includes profile information (nick, about, email),
// verification and status flags (verified, banned, etc.), and engagement statistics
// (subscribers, subscriptions, posts, etc.).
type User struct {
	Email            string `json:"email"`
	SafeMode         bool   `json:"safe_mode"`
	OriginalNick     string `json:"original_nick"`
	MessagingPrivacy string `json:"messaging_privacy_status"`

	ID    string `json:"id"`
	Nick  string `json:"nick"`
	About string `json:"about"`

	IsAvailableForChat    bool `json:"is_available_for_chat"`
	IsBanned              bool `json:"is_banned"`
	IsDeleted             bool `json:"is_deleted"`
	IsModerator           bool `json:"is_moderator"`
	IsVerified            bool `json:"is_verified"`
	IsInSubscribers       bool `json:"is_in_subscribers"`
	IsInSubscriptions     bool `json:"is_in_subscriptions"`
	IsSubscribedToUpdates bool `json:"is_subscribed_to_updates"`
	IsPrivate             bool `json:"is_private"`

	Num struct {
		Subscriptions int `json:"subscriptions"`
		Subscribers   int `json:"subscribers"`
		TotalPosts    int `json:"total_posts"`
		Created       int `json:"created"`
		Featured      int `json:"featured"`
		TotalSmiles   int `json:"total_smiles"`
		Achievements  int `json:"achievements"`
	} `json:"num"`
}

// GetUser fetches a single user given a composed request (e.g. compose.UserAccount() for the authenticated user).
func (client *Client) GetUser(ctx context.Context, desc compose.Request) (*User, error) {
	user := new(struct {
		Data User `json:"data"`
	})

	err := client.RequestJSON(ctx, desc, user)
	return &user.Data, err
}

// UsersEnvelope is the response envelope for user feeds whose page lives at
// data.users (content smiles, content republishers, user subscribers,
// subscriptions). Hand it to [FetchPage]/[Iter] as E. These endpoints return a
// reduced projection of User (no email/privacy fields); absent fields are simply
// zero-valued.
type UsersEnvelope struct {
	Data struct {
		Users Page[User] `json:"users"`
	} `json:"data"`
}

func (e UsersEnvelope) page() Page[User] { return e.Data.Users }

// IterSubscribers returns a channel that yields users who follow the user (identified by ID).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterSubscribers(ctx context.Context, id string) <-chan Result[*User] {
	return Iter[UsersEnvelope](ctx, client, compose.Subscribers(id))
}

// IterSubscriptions returns a channel that yields users followed by the user (identified by ID).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterSubscriptions(ctx context.Context, id string) <-chan Result[*User] {
	return Iter[UsersEnvelope](ctx, client, compose.Subscriptions(id))
}

// GetUsers executes a chat RPC call and unmarshals the result as a list of users.
//
// The desc argument is an opaque turnpike.Call — construct it with a builder
// from the [compose] package that resolves to a user list:
//
//   - [compose.Contacts] — the authenticated user's chat contacts
//   - [compose.SearchContacts] — contacts filtered by a query string
//   - [compose.Operators] — operators of a given channel
//
// Example (list your first 50 contacts):
//
//	users, err := chat.GetUsers(ctx, compose.Contacts(50))
//	if err != nil {
//		return err
//	}
//
// Example (list the operators of a channel):
//
//	ops, err := chat.GetUsers(ctx, compose.Operators("chat.gamers"))
func (chat *Chat) GetUsers(ctx context.Context, desc turnpike.Call) ([]*User, error) {
	output := new(struct {
		Users []*User `json:"users"`
	})

	err := chat.Call(ctx, desc, output)
	return output.Users, err
}
