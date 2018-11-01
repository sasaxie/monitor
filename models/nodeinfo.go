package models

type NodeInfo struct {
	Address              string
	NowBlockNum          int64
	NowBlockHash         string
	LastSolidityBlockNum int64
	RequestTimeConsuming int64
	Message              string
	TotalTransaction     int64
}

type NodeInfoDetail struct {
	ActiveConnectCount  int64
	BeginSyncNum        int64
	Block               string
	CheatWitnessInfoMap map[string]string
	ConfigNodeInfo      *ConfigNodeInfo
	CurrentConnectCount int64
	MachineInfo         *MachineInfo
	PassiveConnectCount int64
	PeerList            []*Peer
	SolidityBlock       string
	TotalFlow           int64
}

type ConfigNodeInfo struct {
	ActiveNodeSize           int64
	AllowCreationOfContracts int64
	BackupListenPort         int64
	BackupMemberSize         int64
	BackupPriority           int64
	CodeVersion              string
	DbVersion                int64
	DiscoverEnable           bool
	ListenPort               int64
	MaxConnectCount          int64
	MaxTimeRatio             float64
	MinParticipationRate     float64
	MinTimeRatio             float64
	P2pVersion               string
	PassiveNodeSize          int64
	SameIpMaxConnectCount    int64
	SendNodeSize             int64
	SupportConstant          bool
}

type MachineInfo struct {
	CpuCount               int64
	CpuRate                float64
	DeadLockThreadCount    int64
	DeadLockThreadInfoList []interface{}
	FreeMemory             int64
	JavaVersion            string
	JvmFreeMemory          int64
	JvmTotalMemoery        int64
	MemoryDescInfoList     []*MemoryDescInfo
	OsName                 string
	ProcessCpuRate         float64
	ThreadCount            int64
	TotalMemory            int64
}

type MemoryDescInfo struct {
	InitSize int64
	MaxSize  int64
	Name     string
	UseRate  float64
	UseSize  int64
}

type Peer struct {
	Active                  bool
	AvgLatency              float64
	BlockInPorcSize         int64
	ConnectTime             int64
	DisconnectTimes         int64
	HeadBlockTimeWeBothHave int64
	HeadBlockWeBothHave     string
	Host                    string
	InFlow                  int64
	LastBlockUpdateTime     int64
	LastSyncBlock           string
	LocalDisconnectReason   string
	NeedSyncFromPeer        bool
	NeedSyncFromUs          bool
	NodeCount               int64
	NodeId                  string
	Port                    int64
	RemainNum               int64
	RemoteDisconnectReason  string
	Score                   int64
	SyncBlockRequestedSize  int64
	SyncFlag                bool
	SyncToFetchSize         int64
	SyncToFetchSizePeekNum  int64
	UnFetchSynNum           int64
}
