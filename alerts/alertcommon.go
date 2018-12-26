package alerts

type Node struct {
	Ip       string
	GrpcPort int
	HttpPort int
	Type     string
	Tag      string
}
