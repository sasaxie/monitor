package parser

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strings"
)

func NilParser(data []byte) (interface{}, error) {
	logs.Debug("nil parsing")
	return nil, nil
}

type Block struct {
	BlockHeader *BlockHeader `json:"block_header"`
}

type BlockHeader struct {
	RawData *RawData `json:"raw_data"`
}

type RawData struct {
	Number int64 `json:"number"`
}

func GetNowBlockParser(data []byte) (interface{}, error) {
	logs.Debug("GetNowBlockParser parsing")
	var block Block
	err := json.Unmarshal(data, &block)

	if block.BlockHeader != nil && block.BlockHeader.RawData != nil {
		logs.Debug(fmt.Sprintf("GetNowBlockParser got block: #%d",
			block.BlockHeader.RawData.Number))
	}

	return block, err
}

type WitnessList struct {
	Witnesses []*Witness `json:"witnesses"`
}

type Witness struct {
	Address     string `json:"address"`
	Url         string `json:"url"`
	TotalMissed int64  `json:"totalMissed"`
	IsJobs      bool   `json:"isJobs"`
}

func ListWitnessesParser(data []byte) (interface{}, error) {
	logs.Debug("ListWitnessesParser parsing")
	var witnesses WitnessList
	err := json.Unmarshal(data, &witnesses)

	if witnesses.Witnesses != nil {
		logs.Debug("ListWitnessesParser got", len(witnesses.Witnesses), "witness")
	}

	return witnesses, err
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

func GetNodeInfoParser(data []byte) (interface{}, error) {
	logs.Debug("GetNodeInfoParser parsing")
	var nodeInfoDetail NodeInfoDetail

	err := json.Unmarshal(data, &nodeInfoDetail)

	return nodeInfoDetail, err
}

type ChainParameters map[string]interface{}

func GetChainParametersParser(data []byte) (interface{}, error) {
	logs.Debug("GetChainParametersParser parsing")

	var origin interface{}
	var chainParameters ChainParameters
	chainParameters = make(map[string]interface{})

	err := json.Unmarshal(data, &origin)
	if err != nil {
		return origin, err
	}

	v := origin.(map[string]interface{})
	for k, vv := range v {
		logs.Info(k, vv)
		vvv := vv.([]interface{})

		for _, vvvv := range vvv {
			vvvvv := vvvv.(map[string]interface{})

			key := ""
			var value int64 = 0
			for kkkkkk, vvvvvv := range vvvvv {
				if strings.EqualFold(kkkkkk, "key") {
					key = vvvvvv.(string)
				} else if strings.EqualFold(kkkkkk, "value") {
					value = int64(vvvvvv.(float64))
				}
			}

			chainParameters[key] = value
		}
	}

	return chainParameters, nil
}
