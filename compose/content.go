package compose

func ContentByID(id string) Request {
	return get("/content/"+id, nil)
}
