package compose

// Comments composes a request for comments on a content item with pagination.
func Comments(id string, limit int, page Page[string]) Request {
	return get("/content/"+id+"/comments", feedParams(limit, page))
}

// Replies composes a request for replies to a comment with pagination.
func Replies(cid, id string, limit int, page Page[string]) Request {
	return get("/content/"+cid+"/comments/"+id+"/replies", feedParams(limit, page))
}
