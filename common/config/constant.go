package config

//=========== Monitor Info ===========//
const MonitorVersion = "v1.0.0"

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
const InfluxDBTagMemoryDescInfoName = "memory_desc_info_name"
const InfluxDBPointNameNodeStatus = "node_status"
const InfluxDBPointNameWitness = "witness"
const InfluxDBPointNameNodeInfoDetail = "node_info_detail"
const InfluxDBFieldNowBlockNum = "NowBlockNum"
const InfluxDBFieldPing = "ping"
const InfluxDBFieldLastSolidityBlockNum = "LastSolidityBlockNum"

const InfluxDBFieldActiveConnectCount = "ActiveConnectCount"
const InfluxDBFieldBeginSyncNum = "BeginSyncNum"
const InfluxDBFieldBlockNum = "BlockNum"
const InfluxDBFieldBlockID = "BlockID"

const InfluxDBFieldActiveNodeSize = "ActiveNodeSize"
const InfluxDBFieldAllowCreationOfContracts = "AllowCreationOfContracts"
const InfluxDBFieldBackupListenPort = "BackupListenPort"
const InfluxDBFieldBackupMemberSize = "BackupMemberSize"
const InfluxDBFieldBackupPriority = "BackupPriority"
const InfluxDBFieldCodeVersion = "CodeVersion"
const InfluxDBFieldDbVersion = "DbVersion"
const InfluxDBFieldDiscoverEnable = "DiscoverEnable"
const InfluxDBFieldListenPort = "ListenPort"
const InfluxDBFieldMaxConnectCount = "MaxConnectCount"
const InfluxDBFieldMaxTimeRatio = "MaxTimeRatio"
const InfluxDBFieldMinParticipationRate = "MinParticipationRate"
const InfluxDBFieldMinTimeRatio = "MinTimeRatio"
const InfluxDBFieldP2pVersion = "P2pVersion"
const InfluxDBFieldPassiveNodeSize = "PassiveNodeSize"
const InfluxDBFieldSameIpMaxConnectCount = "SameIpMaxConnectCount"
const InfluxDBFieldSendNodeSize = "SendNodeSize"
const InfluxDBFieldSupportConstant = "SupportConstant"

const InfluxDBFieldCurrentConnectCount = "CurrentConnectCount"

const InfluxDBFieldCpuCount = "CpuCount"
const InfluxDBFieldCpuRate = "CpuRate"
const InfluxDBFieldDeadLockThreadCount = "DeadLockThreadCount"
const InfluxDBFieldFreeMemory = "FreeMemory"
const InfluxDBFieldJavaVersion = "JavaVersion"
const InfluxDBFieldJvmFreeMemory = "JvmFreeMemory"
const InfluxDBFieldJvmTotalMemoery = "JvmTotalMemoery"

const InfluxDBFieldMemoryDescInfoInitSize = "InitSize"
const InfluxDBFieldMemoryDescInfoMaxSize = "MaxSize"
const InfluxDBFieldMemoryDescInfoUseRate = "UseRate"
const InfluxDBFieldMemoryDescInfoUseSize = "UseSize"

const InfluxDBFieldOsName = "OsName"
const InfluxDBFieldProcessCpuRate = "ProcessCpuRate"
const InfluxDBFieldThreadCount = "ThreadCount"
const InfluxDBFieldTotalMemory = "TotalMemory"

const InfluxDBFieldPassiveConnectCount = "PassiveConnectCount"

const InfluxDBFieldSolidityBlockNum = "SolidityBlockNum"
const InfluxDBFieldSolidityBlockID = "SolidityBlockID"
const InfluxDBFieldTotalFlow = "TotalFlow"

//=========== gRPC ===========//
const GRPCDefaultPort = "50051"
