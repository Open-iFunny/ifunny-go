package ifunny

// Authored represents an iFunny API resource that was created by a user.
type Authored interface {
	AuthorID() string
	AuthorNick() string
}

// AuthorID returns the ID of the user who authored this comment.
func (c *Comment) AuthorID() string { return c.User.ID }

// AuthorNick returns the nick of the user who authored this comment.
func (c *Comment) AuthorNick() string { return c.User.Nick }

// AuthorID returns the ID of the user who authored this content.
func (c *Content) AuthorID() string { return c.Creator.ID }

// AuthorNick returns the nick of the user who authored this content.
func (c *Content) AuthorNick() string { return c.Creator.Nick }

// AuthorID returns the ID of the user who authored this chat event.
func (e *ChatEvent) AuthorID() string { return e.User.ID }

// AuthorNick returns the nick of the user who authored this chat event.
func (e *ChatEvent) AuthorNick() string { return e.User.Nick }
