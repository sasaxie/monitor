package alerts

type Alerter interface {
	Load()
	Start()
	Alert()
}

type Node struct {
	Ip       string
	GrpcPort int
	HttpPort int
	Type     string
	TagName  string
}
