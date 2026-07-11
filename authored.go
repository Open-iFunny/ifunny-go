package ifunny

// Authored represents an iFunny API resource that was created by a user.
type Authored interface {
	AuthorID() string
}

// AuthorID returns the ID of the user who authored this comment.
func (c *Comment) AuthorID() string {
	return c.User.ID
}

// AuthorID returns the ID of the user who authored this content.
func (c *Content) AuthorID() string {
	return c.Creator.ID
}

// AuthorID returns the ID of the user who authored this chat event.
func (e *ChatEvent) AuthorID() string {
	return e.User.ID
}
