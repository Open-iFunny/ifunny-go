package ifunny

import (
	"github.com/gastrodon/turnpike"
	"github.com/open-ifunny/ifunny-go/compose"
)

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

func (client *Client) GetUser(desc compose.Request) (*User, error) {
	user := new(struct {
		Data User `json:"data"`
	})

	err := client.RequestJSON(desc, user)
	return &user.Data, err
}

// GetUsersPage fetches a page of any endpoint whose envelope is
// data.users: content smiles, content republishers, user subscribers, and
// user subscriptions. These endpoints return a reduced projection of User
// (no email/privacy fields); absent fields are simply zero-valued.
func (client *Client) GetUsersPage(request compose.Request) (*Page[User], error) {
	users := new(struct {
		Data struct {
			Users Page[User] `json:"users"`
		} `json:"data"`
	})

	err := client.RequestJSON(request, users)
	return &users.Data.Users, err
}

func (client *Client) IterSubscribers(id string) <-chan Result[*User] {
	return iterFrom(client, func(limit int, page compose.Page[string]) compose.Request {
		return compose.Subscribers(id, limit, page)
	}, client.GetUsersPage)
}

func (client *Client) IterSubscriptions(id string) <-chan Result[*User] {
	return iterFrom(client, func(limit int, page compose.Page[string]) compose.Request {
		return compose.Subscriptions(id, limit, page)
	}, client.GetUsersPage)
}

func (chat *Chat) GetUsers(desc turnpike.Call) ([]*User, error) {
	output := new(struct {
		Users []*User `json:"users"`
	})

	err := chat.Call(desc, output)
	return output.Users, err
}
