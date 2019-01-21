package config

//=========== Monitor Info ===========//
const MonitorVersion = "v3.0.3"

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
	Unknown
)

func NewNodeType(s string) NodeType {
	switch s {
	case "full_node":
		return FullNode
	case "mti_full_node":
		return MtiFullNode
	case "witness_node":
		return WitnessNode
	case "sr_witness_node":
		return SRWitnessNode
	case "sr_witness_b_node":
		return SRWitnessBNode
	case "gr_witness_node":
		return GRWitnessNode
	case "solidity_node":
		return SolidityNode
	default:
		return Unknown
	}
}

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

func (n NodeType) GetApiPathByNodeType() string {
	switch n {
	case FullNode:
		fallthrough
	case MtiFullNode:
		fallthrough
	case WitnessNode:
		fallthrough
	case SRWitnessNode:
		fallthrough
	case SRWitnessBNode:
		fallthrough
	case GRWitnessNode:
		return "wallet"
	case SolidityNode:
		return "walletsolidity"
	default:
		return "Unknown"
	}
}
