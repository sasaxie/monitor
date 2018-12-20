package config

//=========== Monitor Info ===========//
const MonitorVersion = "v2.1.1"

//=========== Node type ===========//
type NodeType int

const (
	FullNode NodeType = iota
	MtiFullNode
	WitnessNode
	SRWitnessNode
	SRWitnessBNode
	GRWitnessNode
	SolidityNode
)

func (n NodeType) String() string {
	switch n {
	case FullNode:
		return "full_node"
	case MtiFullNode:
		return "mti_full_node"
	case WitnessNode:
		return "witness_node"
	case SRWitnessNode:
		return "sr_witness_node"
	case SRWitnessBNode:
		return "sr_witness_b_node"
	case GRWitnessNode:
		return "gr_witness_node"
	case SolidityNode:
		return "solidity_node"
	default:
		return "Unknown"
	}
}
