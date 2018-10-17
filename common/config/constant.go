package config

//=========== Node type ===========//
type NodeType int

const (
	FullNode NodeType = iota
	SolidityNode
)

func (n NodeType) String() string {
	switch n {
	case FullNode:
		return "full_node"
	case SolidityNode:
		return "solidity_node"
	default:
		return "Unknown"
	}
}

//=========== InfluxDB ===========//
const InfluxDBTagNode = "node"
const InfluxDBPointNameNodeStatus = "node_status"
const InfluxDBPointNameWitness = "witness"
const InfluxDBFieldNowBlockNum = "NowBlockNum"
const InfluxDBFieldPing = "ping"
const InfluxDBFieldLastSolidityBlockNum = "LastSolidityBlockNum"
