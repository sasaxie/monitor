package collector

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
	"io/ioutil"
	"net/http"
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

type GetNodeInfoRequest struct {
	Common
}

func init() {
	Collectors = append(Collectors, new(GetNodeInfoRequest))
}

func (g *GetNodeInfoRequest) Collect() {
	if !g.HasInitNodes {
		g.initNodes()
		g.HasInitNodes = true
	}

	g.start()
}

func (g *GetNodeInfoRequest) initNodes() {
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

func (g *GetNodeInfoRequest) start() {
	if g.Nodes == nil || len(g.Nodes) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Nodes))
	for _, node := range g.Nodes {
		go g.request(node, &wg)
	}

	wg.Wait()
}

func (g *GetNodeInfoRequest) request(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := http.Get(node.CollectionUrl)

	if err != nil {
		logs.Debug("(", node.CollectionUrl, ")", "[http get]", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logs.Warn("get node info request (", node.CollectionUrl,
			") response status code",
			response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[read body]", err)
		return
	}

	var nodeInfoDetail models.NodeInfoDetail
	err = json.Unmarshal(body, &nodeInfoDetail)

	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[json unmarshal]", err)
		return
	}

	blockNum, blockID := getBlockNumAndId(nodeInfoDetail.Block)

	solidityBlockNum, solidityBlockID := getBlockNumAndId(nodeInfoDetail.SolidityBlock)

	timeNow := time.Now()

	nodeInfoDetailTags := map[string]string{
		influxDBTagGetNodeInfoNode:    node.Node,
		influxDBTagGetNodeInfoType:    node.Type,
		influxDBTagGetNodeInfoTagName: node.TagName,
	}

	cheatWitnessInfo := getCheatWitnessInfoStr(nodeInfoDetail.CheatWitnessInfoMap)

	nodeInfoDetailFields := map[string]interface{}{
		influxDBFieldGetNodeInfoNode:    node.Node,
		influxDBFieldGetNodeInfoType:    node.Type,
		influxDBFieldGetNodeInfoTagName: node.TagName,

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

	influxdb.Client.WriteByTime(
		influxDBPointNameGetNodeInfo,
		nodeInfoDetailTags,
		nodeInfoDetailFields,
		timeNow)

	for _, v := range nodeInfoDetail.MachineInfo.MemoryDescInfoList {
		t := map[string]string{
			influxDBTagGetNodeInfoNode:               node.Node,
			influxDBTagGetNodeInfoMemoryDescInfoName: v.Name,
		}
		f := map[string]interface{}{
			influxDBFieldGetNodeInfoMemoryDescInfoInitSize: v.InitSize,
			influxDBFieldGetNodeInfoMemoryDescInfoMaxSize:  v.MaxSize,
			influxDBFieldGetNodeInfoMemoryDescInfoUseRate:  v.UseRate * 100,
			influxDBFieldGetNodeInfoMemoryDescInfoUseSize:  v.UseSize,
		}

		influxdb.Client.WriteByTime(
			influxDBPointNameGetNodeInfo,
			t,
			f,
			timeNow)
	}

	for _, p := range nodeInfoDetail.PeerList {
		hbn, hbi := getBlockNumAndId(p.HeadBlockWeBothHave)

		ln, li := getBlockNumAndId(p.LastSyncBlock)

		t := map[string]string{
			influxDBTagGetNodeInfoNode: node.Node,
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

		influxdb.Client.WriteByTime(
			influxDBPointNameGetNodeInfoPeerInfo,
			t,
			f,
			timeNow)
	}
}

func getBlockNumAndId(blockStr string) (int64, string) {
	var num int64 = 0
	var id = ""
	var err error
	if len(blockStr) > 0 && !strings.EqualFold(blockStr, "") {
		strs := strings.Split(blockStr, ",")
		if len(strs) > 0 {
			numStrs := strings.Split(strs[0], ":")
			if len(numStrs) > 0 {
				num, err = strconv.ParseInt(numStrs[1], 10, 64)
				if err != nil {
					logs.Warn(err)
				}
			}

			idStrs := strings.Split(strs[1], ":")
			if len(idStrs) > 0 {
				id = idStrs[1]
			}
		}
	}

	return num, id
}

func getCheatWitnessInfoStr(cheatWitnessInfoMap map[string]string) string {
	cheatWitnessInfo := ""
	for k, v := range cheatWitnessInfoMap {
		cheatWitnessInfo += k + ":" + v + " "
	}

	return cheatWitnessInfo
}
