package compose

func getFeed(path, next, prev string) Request {
	return get(path, nil) // TODO this feels incomplete
}

func FeedCollective(prev, next string) Request {
	return Request{Path: "/feeds/collective", Method: "POST"}
}

func FeedFeatures(prev, next string) Request {
	return getFeed("/feeds/collective", prev, next)
}

func FeedHome(prev, next string) Request {
	return getFeed("/timelines/home", prev, next)
}

func UserTimelineByID(id, prev, next string) Request {
	return getFeed("/timelines/users/"+id, prev, next)
}

func UserTimelineByNick(nick, prev, next string) Request {
	return UserTimelineByID("by_nick/"+nick, prev, next)
}
