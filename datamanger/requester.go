package datamanger

type Requester interface {
	Load()
	Request()
	Save2db()
}

type RequestCommon struct {
	Parameters []*Parameter
}

type Parameter struct {
	RequestUrl string

	// InfluxDB tags
	Node    string
	Type    string
	TagName string
}
