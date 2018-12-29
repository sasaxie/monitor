package alerts

type Node struct {
	Ip       string
	GrpcPort int
	HttpPort int
	Type     string
	TagName  string
}
