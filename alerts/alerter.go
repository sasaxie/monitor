package alerts

const (
	Internal1min int64 = 1000 * 60 * 1
	Internal5min int64 = 1000 * 60 * 5
)

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
