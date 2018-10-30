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

//------------ Tag ------------//
const InfluxDBTagNode = "node"
const InfluxDBTagMemoryDescInfoName = "memory_desc_info_name"
const InfluxDBTagPeer = "peer"

//------------ Point ------------//
const InfluxDBPointNameNodeStatus = "node_status"
const InfluxDBPointNameWitness = "witness"
const InfluxDBPointNameNodeInfoDetail = "node_info_detail"
const InfluxDBPointNamePeerInfo = "peer_info"

//------------ node_status Field ------------//
const InfluxDBFieldNowBlockNum = "NowBlockNum"
const InfluxDBFieldPing = "ping"
const InfluxDBFieldLastSolidityBlockNum = "LastSolidityBlockNum"

//------------ node_info_detail Field ------------//

// Basic information
const InfluxDBFieldActiveConnectCount = "ActiveConnectCount"
const InfluxDBFieldBeginSyncNum = "BeginSyncNum"
const InfluxDBFieldBlockNum = "BlockNum"
const InfluxDBFieldBlockID = "BlockID"
const InfluxDBFieldCurrentConnectCount = "CurrentConnectCount"
const InfluxDBFieldPassiveConnectCount = "PassiveConnectCount"
const InfluxDBFieldSolidityBlockNum = "SolidityBlockNum"
const InfluxDBFieldSolidityBlockID = "SolidityBlockID"
const InfluxDBFieldTotalFlow = "TotalFlow"

// Configuration information
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

// System information
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

// peer
const InfluxDBFieldPeerActive = "Active"
const InfluxDBFieldPeerAvgLatency = "AvgLatency"
const InfluxDBFieldPeerBlockInPorcSize = "BlockInPorcSize"
const InfluxDBFieldPeerConnectTime = "ConnectTime"
const InfluxDBFieldPeerDisconnectTimes = "DisconnectTimes"
const InfluxDBFieldPeerHeadBlockTimeWeBothHave = "HeadBlockTimeWeBothHave"
const InfluxDBFieldPeerHeadBlockWeBothHaveNum = "HeadBlockWeBothHaveNum"
const InfluxDBFieldPeerHeadBlockWeBothHaveID = "HeadBlockWeBothHaveID"
const InfluxDBFieldPeerHost = "Host"
const InfluxDBFieldPeerInFlow = "InFlow"
const InfluxDBFieldPeerLastBlockUpdateTime = "LastBlockUpdateTime"
const InfluxDBFieldPeerLastSyncBlockNum = "LastSyncBlockNum"
const InfluxDBFieldPeerLastSyncBlockID = "LastSyncBlockID"
const InfluxDBFieldPeerLocalDisconnectReason = "LocalDisconnectReason"
const InfluxDBFieldPeerNeedSyncFromPeer = "NeedSyncFromPeer"
const InfluxDBFieldPeerNeedSyncFromUs = "NeedSyncFromUs"
const InfluxDBFieldPeerNodeCount = "NodeCount"
const InfluxDBFieldPeerNodeID = "NodeID"
const InfluxDBFieldPeerPort = "Port"
const InfluxDBFieldPeerRemainNum = "RemainNum"
const InfluxDBFieldPeerRemoteDisconnectReason = "RemoteDisconnectReason"
const InfluxDBFieldPeerScore = "Score"
const InfluxDBFieldPeerSyncBlockRequestedSize = "SyncBlockRequestedSize"
const InfluxDBFieldPeerSyncFlag = "SyncFlag"
const InfluxDBFieldPeerSyncToFetchSize = "SyncToFetchSize"
const InfluxDBFieldPeerSyncToFetchSizePeekNum = "SyncToFetchSizePeekNum"
const InfluxDBFieldPeerUnFetchSynNum = "UnFetchSynNum"

//=========== gRPC ===========//
const GRPCDefaultPort = "50051"
