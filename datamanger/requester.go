package datamanger

type Requester interface {
	Load()
	Request()
	Save2db()
}
