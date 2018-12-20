package datamanger

type RequestCommon struct {
	Parameters []*Parameter
}

type Parameter struct {
	RequestUrl string

	// InfluxDB tags
	Node string
	Type string
	Tag  string
}
