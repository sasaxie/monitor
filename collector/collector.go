package collector

type Collector interface {
	Collect()
}

type Common struct {
	Nodes []*Node

	HasInitNodes bool
}

type Node struct {
	CollectionUrl string

	// InfluxDB tags
	Node    string
	Type    string
	TagName string
}
