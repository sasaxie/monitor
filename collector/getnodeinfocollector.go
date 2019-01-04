package collector

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	urlTemplateGetNodeInfo = "http://%s:%d/%s/getnodeinfo"

	influxDBTagGetNodeInfoNode    = "node"
	influxDBTagGetNodeInfoType    = "type"
	influxDBTagGetNodeInfoTagName = "tag"
	influxDBTagGetNodeInfoPeer    = "peer"

	influxDBFieldGetNodeInfoNode    = "Node"
	influxDBFieldGetNodeInfoType    = "Type"
	influxDBFieldGetNodeInfoTagName = "TagName"

	// Basic information
	influxDBFieldGetNodeInfoActiveConnectCount  = "ActiveConnectCount"
	influxDBFieldGetNodeInfoBeginSyncNum        = "BeginSyncNum"
	influxDBFieldGetNodeInfoBlockNum            = "BlockNum"
	influxDBFieldGetNodeInfoBlockID             = "BlockID"
	influxDBFieldGetNodeInfoCurrentConnectCount = "CurrentConnectCount"
	influxDBFieldGetNodeInfoPassiveConnectCount = "PassiveConnectCount"
	influxDBFieldGetNodeInfoSolidityBlockNum    = "SolidityBlockNum"
	influxDBFieldGetNodeInfoSolidityBlockID     = "SolidityBlockID"
	influxDBFieldGetNodeInfoTotalFlow           = "TotalFlow"

	// Cheat witness
	influxDBFieldGetNodeInfoCheatWitnessInfoMap = "CheatWitnessInfoMap"

	// Configuration information
	influxDBFieldGetNodeInfoActiveNodeSize           = "ActiveNodeSize"
	influxDBFieldGetNodeInfoAllowCreationOfContracts = "AllowCreationOfContracts"
	influxDBFieldGetNodeInfoBackupListenPort         = "BackupListenPort"
	influxDBFieldGetNodeInfoBackupMemberSize         = "BackupMemberSize"
	influxDBFieldGetNodeInfoBackupPriority           = "BackupPriority"
	influxDBFieldGetNodeInfoCodeVersion              = "CodeVersion"
	influxDBFieldGetNodeInfoDbVersion                = "DbVersion"
	influxDBFieldGetNodeInfoDiscoverEnable           = "DiscoverEnable"
	influxDBFieldGetNodeInfoListenPort               = "ListenPort"
	influxDBFieldGetNodeInfoMaxConnectCount          = "MaxConnectCount"
	influxDBFieldGetNodeInfoMaxTimeRatio             = "MaxTimeRatio"
	influxDBFieldGetNodeInfoMinParticipationRate     = "MinParticipationRate"
	influxDBFieldGetNodeInfoMinTimeRatio             = "MinTimeRatio"
	influxDBFieldGetNodeInfoP2pVersion               = "P2pVersion"
	influxDBFieldGetNodeInfoPassiveNodeSize          = "PassiveNodeSize"
	influxDBFieldGetNodeInfoSameIpMaxConnectCount    = "SameIpMaxConnectCount"
	influxDBFieldGetNodeInfoSendNodeSize             = "SendNodeSize"
	influxDBFieldGetNodeInfoSupportConstant          = "SupportConstant"

	// System information
	influxDBFieldGetNodeInfoCpuCount            = "CpuCount"
	influxDBFieldGetNodeInfoCpuRate             = "CpuRate"
	influxDBFieldGetNodeInfoDeadLockThreadCount = "DeadLockThreadCount"
	influxDBFieldGetNodeInfoFreeMemory          = "FreeMemory"
	influxDBFieldGetNodeInfoJavaVersion         = "JavaVersion"
	influxDBFieldGetNodeInfoJvmFreeMemory       = "JvmFreeMemory"
	influxDBFieldGetNodeInfoJvmTotalMemoery     = "JvmTotalMemoery"

	influxDBFieldGetNodeInfoMemoryDescInfoInitSize = "InitSize"
	influxDBFieldGetNodeInfoMemoryDescInfoMaxSize  = "MaxSize"
	influxDBFieldGetNodeInfoMemoryDescInfoUseRate  = "UseRate"
	influxDBFieldGetNodeInfoMemoryDescInfoUseSize  = "UseSize"

	influxDBFieldGetNodeInfoOsName         = "OsName"
	influxDBFieldGetNodeInfoProcessCpuRate = "ProcessCpuRate"
	influxDBFieldGetNodeInfoThreadCount    = "ThreadCount"
	influxDBFieldGetNodeInfoTotalMemory    = "TotalMemory"

	// peer
	influxDBFieldGetNodeInfoPeerActive                  = "Active"
	influxDBFieldGetNodeInfoPeerAvgLatency              = "AvgLatency"
	influxDBFieldGetNodeInfoPeerBlockInPorcSize         = "BlockInPorcSize"
	influxDBFieldGetNodeInfoPeerConnectTime             = "ConnectTime"
	influxDBFieldGetNodeInfoPeerDisconnectTimes         = "DisconnectTimes"
	influxDBFieldGetNodeInfoPeerHeadBlockTimeWeBothHave = "HeadBlockTimeWeBothHave"
	influxDBFieldGetNodeInfoPeerHeadBlockWeBothHaveNum  = "HeadBlockWeBothHaveNum"
	influxDBFieldGetNodeInfoPeerHeadBlockWeBothHaveID   = "HeadBlockWeBothHaveID"
	influxDBFieldGetNodeInfoPeerHost                    = "Host"
	influxDBFieldGetNodeInfoPeerInFlow                  = "InFlow"
	influxDBFieldGetNodeInfoPeerLastBlockUpdateTime     = "LastBlockUpdateTime"
	influxDBFieldGetNodeInfoPeerLastSyncBlockNum        = "LastSyncBlockNum"
	influxDBFieldGetNodeInfoPeerLastSyncBlockID         = "LastSyncBlockID"
	influxDBFieldGetNodeInfoPeerLocalDisconnectReason   = "LocalDisconnectReason"
	influxDBFieldGetNodeInfoPeerNeedSyncFromPeer        = "NeedSyncFromPeer"
	influxDBFieldGetNodeInfoPeerNeedSyncFromUs          = "NeedSyncFromUs"
	influxDBFieldGetNodeInfoPeerNodeCount               = "NodeCount"
	influxDBFieldGetNodeInfoPeerNodeID                  = "NodeID"
	influxDBFieldGetNodeInfoPeerPort                    = "Port"
	influxDBFieldGetNodeInfoPeerRemainNum               = "RemainNum"
	influxDBFieldGetNodeInfoPeerRemoteDisconnectReason  = "RemoteDisconnectReason"
	influxDBFieldGetNodeInfoPeerScore                   = "Score"
	influxDBFieldGetNodeInfoPeerSyncBlockRequestedSize  = "SyncBlockRequestedSize"
	influxDBFieldGetNodeInfoPeerSyncFlag                = "SyncFlag"
	influxDBFieldGetNodeInfoPeerSyncToFetchSize         = "SyncToFetchSize"
	influxDBFieldGetNodeInfoPeerSyncToFetchSizePeekNum  = "SyncToFetchSizePeekNum"
	influxDBFieldGetNodeInfoPeerUnFetchSynNum           = "UnFetchSynNum"

	influxDBPointNameGetNodeInfoPeerInfo     = "api_peer_info"
	influxDBPointNameGetNodeInfo             = "api_get_node_info"
	influxDBTagGetNodeInfoMemoryDescInfoName = "api_memory_desc_info_name"
)

var Collectors = make([]Collector, 0)

type GetNodeInfoCollector struct {
	Common
}

func init() {
	Collectors = append(Collectors, new(GetNodeInfoCollector))
}

func (g *GetNodeInfoCollector) Collect() {
	g.init()

	g.collect()
}

func (g *GetNodeInfoCollector) init() {
	if !g.HasInit {
		g.initNodes()
		g.HasInit = true
		logs.Info("init GetNodeInfoCollector")
	}
}

func (g *GetNodeInfoCollector) initNodes() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("get node info request load() error")
	}

	if g.Nodes == nil {
		g.Nodes = make([]*Node, 0)
	}

	for _, node := range models.NodeList.Addresses {
		if strings.EqualFold(node.Type, config.SolidityNode.String()) {
			continue
		}

		n := new(Node)
		n.CollectionUrl = fmt.Sprintf(
			urlTemplateGetNodeInfo,
			node.Ip,
			node.HttpPort,
			config.NewNodeType(node.Type).GetApiPathByNodeType())
		n.Node = fmt.Sprintf("%s:%d", node.Ip, node.HttpPort)
		n.Type = node.Type
		n.TagName = node.TagName

		g.Nodes = append(g.Nodes, n)
	}

	logs.Info(
		"get node info request load() success, node size:",
		len(g.Nodes),
	)
}

func (g *GetNodeInfoCollector) collect() {
	if g.Nodes == nil || len(g.Nodes) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Nodes))
	for _, node := range g.Nodes {
		go g.collectByNode(node, &wg)
	}

	wg.Wait()
}

func (g *GetNodeInfoCollector) collectByNode(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := fetch(node.CollectionUrl)
	if err != nil {
		logs.Debug(err)
		return
	}

	var nodeInfoDetail models.NodeInfoDetail
	err = json.Unmarshal(data, &nodeInfoDetail)

	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[json unmarshal]", err)
		return
	}

	timeNow := time.Now()

	g.saveNodeInfoDetail(nodeInfoDetail, node.Node, node.Type, node.TagName, timeNow)
	g.saveMemoryDescInfoList(nodeInfoDetail, node.Node, timeNow)
	g.savePeerList(nodeInfoDetail, node.Node, timeNow)
}

func (g *GetNodeInfoCollector) saveNodeInfoDetail(
	nodeInfoDetail models.NodeInfoDetail,
	nodeHost, nodeType, nodeTagName string,
	timeNow time.Time) {
	blockNum, blockID := g.getBlockNumAndId(nodeInfoDetail.Block)

	solidityBlockNum, solidityBlockID := g.getBlockNumAndId(nodeInfoDetail.
		SolidityBlock)

	nodeInfoDetailTags := map[string]string{
		influxDBTagGetNodeInfoNode:    nodeHost,
		influxDBTagGetNodeInfoType:    nodeType,
		influxDBTagGetNodeInfoTagName: nodeTagName,
	}

	cheatWitnessInfo := g.getCheatWitnessInfoStr(nodeInfoDetail.
		CheatWitnessInfoMap)

	nodeInfoDetailFields := map[string]interface{}{
		influxDBFieldGetNodeInfoNode:    nodeHost,
		influxDBFieldGetNodeInfoType:    nodeType,
		influxDBFieldGetNodeInfoTagName: nodeTagName,

		influxDBFieldGetNodeInfoActiveConnectCount: nodeInfoDetail.ActiveConnectCount,
		influxDBFieldGetNodeInfoBeginSyncNum:       nodeInfoDetail.BeginSyncNum,
		influxDBFieldGetNodeInfoBlockNum:           blockNum,
		influxDBFieldGetNodeInfoBlockID:            blockID,

		influxDBFieldGetNodeInfoCheatWitnessInfoMap: cheatWitnessInfo,

		influxDBFieldGetNodeInfoActiveNodeSize:           nodeInfoDetail.ConfigNodeInfo.ActiveNodeSize,
		influxDBFieldGetNodeInfoAllowCreationOfContracts: nodeInfoDetail.ConfigNodeInfo.AllowCreationOfContracts,
		influxDBFieldGetNodeInfoBackupListenPort:         nodeInfoDetail.ConfigNodeInfo.BackupListenPort,
		influxDBFieldGetNodeInfoBackupMemberSize:         nodeInfoDetail.ConfigNodeInfo.BackupMemberSize,
		influxDBFieldGetNodeInfoBackupPriority:           nodeInfoDetail.ConfigNodeInfo.BackupPriority,
		influxDBFieldGetNodeInfoCodeVersion:              nodeInfoDetail.ConfigNodeInfo.CodeVersion,
		influxDBFieldGetNodeInfoDbVersion:                nodeInfoDetail.ConfigNodeInfo.DbVersion,
		influxDBFieldGetNodeInfoDiscoverEnable:           nodeInfoDetail.ConfigNodeInfo.DiscoverEnable,
		influxDBFieldGetNodeInfoListenPort:               nodeInfoDetail.ConfigNodeInfo.ListenPort,
		influxDBFieldGetNodeInfoMaxConnectCount:          nodeInfoDetail.ConfigNodeInfo.MaxConnectCount,
		influxDBFieldGetNodeInfoMaxTimeRatio:             nodeInfoDetail.ConfigNodeInfo.MaxTimeRatio,
		influxDBFieldGetNodeInfoMinParticipationRate:     nodeInfoDetail.ConfigNodeInfo.MinParticipationRate,
		influxDBFieldGetNodeInfoMinTimeRatio:             nodeInfoDetail.ConfigNodeInfo.MinTimeRatio,
		influxDBFieldGetNodeInfoP2pVersion:               nodeInfoDetail.ConfigNodeInfo.P2pVersion,
		influxDBFieldGetNodeInfoPassiveNodeSize:          nodeInfoDetail.ConfigNodeInfo.PassiveNodeSize,
		influxDBFieldGetNodeInfoSameIpMaxConnectCount:    nodeInfoDetail.ConfigNodeInfo.SameIpMaxConnectCount,
		influxDBFieldGetNodeInfoSendNodeSize:             nodeInfoDetail.ConfigNodeInfo.SendNodeSize,
		influxDBFieldGetNodeInfoSupportConstant:          nodeInfoDetail.ConfigNodeInfo.SupportConstant,

		influxDBFieldGetNodeInfoCurrentConnectCount: nodeInfoDetail.CurrentConnectCount,

		influxDBFieldGetNodeInfoCpuCount:            nodeInfoDetail.MachineInfo.CpuCount,
		influxDBFieldGetNodeInfoCpuRate:             nodeInfoDetail.MachineInfo.CpuRate * 100,
		influxDBFieldGetNodeInfoDeadLockThreadCount: nodeInfoDetail.MachineInfo.DeadLockThreadCount,
		influxDBFieldGetNodeInfoFreeMemory:          nodeInfoDetail.MachineInfo.FreeMemory,
		influxDBFieldGetNodeInfoJavaVersion:         nodeInfoDetail.MachineInfo.JavaVersion,
		influxDBFieldGetNodeInfoJvmFreeMemory:       nodeInfoDetail.MachineInfo.JvmFreeMemory,
		influxDBFieldGetNodeInfoJvmTotalMemoery:     nodeInfoDetail.MachineInfo.JvmTotalMemoery,

		influxDBFieldGetNodeInfoOsName:         nodeInfoDetail.MachineInfo.OsName,
		influxDBFieldGetNodeInfoProcessCpuRate: nodeInfoDetail.MachineInfo.ProcessCpuRate * 100,
		influxDBFieldGetNodeInfoThreadCount:    nodeInfoDetail.MachineInfo.ThreadCount,
		influxDBFieldGetNodeInfoTotalMemory:    nodeInfoDetail.MachineInfo.TotalMemory,

		influxDBFieldGetNodeInfoPassiveConnectCount: nodeInfoDetail.PassiveConnectCount,

		influxDBFieldGetNodeInfoSolidityBlockNum: solidityBlockNum,
		influxDBFieldGetNodeInfoSolidityBlockID:  solidityBlockID,
		influxDBFieldGetNodeInfoTotalFlow:        nodeInfoDetail.TotalFlow,
	}

	err := influxdb.Client.WriteByTime(
		influxDBPointNameGetNodeInfo,
		nodeInfoDetailTags,
		nodeInfoDetailFields,
		timeNow)

	if err != nil {
		logs.Error(err)
	}
}

func (g *GetNodeInfoCollector) saveMemoryDescInfoList(
	nodeInfoDetail models.NodeInfoDetail,
	nodeHost string,
	timeNow time.Time) {
	for _, v := range nodeInfoDetail.MachineInfo.MemoryDescInfoList {
		t := map[string]string{
			influxDBTagGetNodeInfoNode:               nodeHost,
			influxDBTagGetNodeInfoMemoryDescInfoName: v.Name,
		}
		f := map[string]interface{}{
			influxDBFieldGetNodeInfoMemoryDescInfoInitSize: v.InitSize,
			influxDBFieldGetNodeInfoMemoryDescInfoMaxSize:  v.MaxSize,
			influxDBFieldGetNodeInfoMemoryDescInfoUseRate:  v.UseRate * 100,
			influxDBFieldGetNodeInfoMemoryDescInfoUseSize:  v.UseSize,
		}

		err := influxdb.Client.WriteByTime(
			influxDBPointNameGetNodeInfo,
			t,
			f,
			timeNow)

		if err != nil {
			logs.Error(err)
		}
	}
}

func (g *GetNodeInfoCollector) savePeerList(
	nodeInfoDetail models.NodeInfoDetail,
	nodeHost string,
	timeNow time.Time) {
	for _, p := range nodeInfoDetail.PeerList {
		hbn, hbi := g.getBlockNumAndId(p.HeadBlockWeBothHave)

		ln, li := g.getBlockNumAndId(p.LastSyncBlock)

		t := map[string]string{
			influxDBTagGetNodeInfoNode: nodeHost,
			influxDBTagGetNodeInfoPeer: p.Host,
		}
		f := map[string]interface{}{
			influxDBFieldGetNodeInfoPeerActive:                  p.Active,
			influxDBFieldGetNodeInfoPeerAvgLatency:              p.AvgLatency,
			influxDBFieldGetNodeInfoPeerBlockInPorcSize:         p.BlockInPorcSize,
			influxDBFieldGetNodeInfoPeerConnectTime:             p.ConnectTime,
			influxDBFieldGetNodeInfoPeerDisconnectTimes:         p.DisconnectTimes,
			influxDBFieldGetNodeInfoPeerHeadBlockTimeWeBothHave: p.HeadBlockTimeWeBothHave,
			influxDBFieldGetNodeInfoPeerHeadBlockWeBothHaveNum:  hbn,
			influxDBFieldGetNodeInfoPeerHeadBlockWeBothHaveID:   hbi,
			influxDBFieldGetNodeInfoPeerHost:                    p.Host,
			influxDBFieldGetNodeInfoPeerInFlow:                  p.InFlow,
			influxDBFieldGetNodeInfoPeerLastBlockUpdateTime:     p.LastBlockUpdateTime,
			influxDBFieldGetNodeInfoPeerLastSyncBlockNum:        ln,
			influxDBFieldGetNodeInfoPeerLastSyncBlockID:         li,
			influxDBFieldGetNodeInfoPeerLocalDisconnectReason:   p.LocalDisconnectReason,
			influxDBFieldGetNodeInfoPeerNeedSyncFromPeer:        p.NeedSyncFromPeer,
			influxDBFieldGetNodeInfoPeerNeedSyncFromUs:          p.NeedSyncFromUs,
			influxDBFieldGetNodeInfoPeerNodeCount:               p.NodeCount,
			influxDBFieldGetNodeInfoPeerNodeID:                  p.NodeId,
			influxDBFieldGetNodeInfoPeerPort:                    p.Port,
			influxDBFieldGetNodeInfoPeerRemainNum:               p.RemainNum,
			influxDBFieldGetNodeInfoPeerRemoteDisconnectReason:  p.RemoteDisconnectReason,
			influxDBFieldGetNodeInfoPeerScore:                   p.Score,
			influxDBFieldGetNodeInfoPeerSyncBlockRequestedSize:  p.SyncBlockRequestedSize,
			influxDBFieldGetNodeInfoPeerSyncFlag:                p.SyncFlag,
			influxDBFieldGetNodeInfoPeerSyncToFetchSize:         p.SyncToFetchSize,
			influxDBFieldGetNodeInfoPeerSyncToFetchSizePeekNum:  p.SyncToFetchSizePeekNum,
			influxDBFieldGetNodeInfoPeerUnFetchSynNum:           p.UnFetchSynNum,
		}

		err := influxdb.Client.WriteByTime(
			influxDBPointNameGetNodeInfoPeerInfo,
			t,
			f,
			timeNow)

		if err != nil {
			logs.Error(err)
		}
	}
}

func (g *GetNodeInfoCollector) getBlockNumAndId(blockStr string) (int64, string) {
	var num int64 = 0
	var id = ""
	var err error
	if len(blockStr) > 0 && !strings.EqualFold(blockStr, "") {
		blockSlice := strings.Split(blockStr, ",")
		if len(blockSlice) > 0 {
			numSlice := strings.Split(blockSlice[0], ":")
			if len(numSlice) > 0 {
				num, err = strconv.ParseInt(numSlice[1], 10, 64)
				if err != nil {
					logs.Warn(err)
				}
			}

			idSlice := strings.Split(blockSlice[1], ":")
			if len(idSlice) > 0 {
				id = idSlice[1]
			}
		}
	}

	return num, id
}

func (g *GetNodeInfoCollector) getCheatWitnessInfoStr(
	cheatWitnessInfoMap map[string]string) string {
	cheatWitnessInfo := ""
	for k, v := range cheatWitnessInfoMap {
		cheatWitnessInfo += k + ":" + v + " "
	}

	return cheatWitnessInfo
}
