package datamanger

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

	influxDBTagGetNodeInfoNode = "node"
	influxDBTagGetNodeInfoType = "type"
	influxDBTagGetNodeInfoTag  = "tag"
	influxDBTagGetNodeInfoPeer = "peer"

	influxDBFieldGetNodeInfoNode = "Node"
	influxDBFieldGetNodeInfoType = "Type"
	influxDBFieldGetNodeInfoTag  = "Tag"

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

var Requests = make([]Requester, 0)

type GetNodeInfoRequest struct {
	RequestCommon
}

func init() {
	Requests = append(Requests, new(GetNodeInfoRequest))
}

func (g *GetNodeInfoRequest) Load() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("get node info request load() error")
	}

	if g.Parameters == nil {
		g.Parameters = make([]*Parameter, 0)
	}

	for _, node := range models.NodeList.Addresses {
		if strings.EqualFold(node.Type, config.SolidityNode.String()) {
			continue
		}

		param := new(Parameter)
		param.RequestUrl = fmt.Sprintf(
			urlTemplateGetNodeInfo,
			node.Ip,
			node.HttpPort,
			config.NewNodeType(node.Type).GetApiPathByNodeType())
		param.Node = fmt.Sprintf("%s:%d", node.Ip, node.HttpPort)
		param.Type = node.Type
		param.Tag = node.Tag

		g.Parameters = append(g.Parameters, param)
	}

	logs.Info(
		"get node info request load() success, node size:",
		len(g.Parameters),
	)
}

func (g *GetNodeInfoRequest) Request() {
	if g.Parameters == nil || len(g.Parameters) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Parameters))
	for _, param := range g.Parameters {
		go g.request(param, &wg)
	}

	wg.Wait()
}

func (g *GetNodeInfoRequest) Save2db() {

}

func (g *GetNodeInfoRequest) request(param *Parameter, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := http.Get(param.RequestUrl)

	if err != nil {
		logs.Debug("(", param.RequestUrl, ")", "[http get]", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logs.Warn("get node info request (", param.RequestUrl,
			") response status code",
			response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Warn("(", param.RequestUrl, ") ", "[read body]", err)
		return
	}

	var nodeInfoDetail models.NodeInfoDetail
	err = json.Unmarshal(body, &nodeInfoDetail)

	if err != nil {
		logs.Warn("(", param.RequestUrl, ") ", "[json unmarshal]", err)
		return
	}

	blockNum, blockID := getBlockNumAndId(nodeInfoDetail.Block)

	solidityBlockNum, solidityBlockID := getBlockNumAndId(nodeInfoDetail.SolidityBlock)

	timeNow := time.Now()

	nodeInfoDetailTags := map[string]string{
		influxDBTagGetNodeInfoNode: param.Node,
		influxDBTagGetNodeInfoType: param.Type,
		influxDBTagGetNodeInfoTag:  param.Tag,
	}

	cheatWitnessInfo := getCheatWitnessInfoStr(nodeInfoDetail.CheatWitnessInfoMap)

	nodeInfoDetailFields := map[string]interface{}{
		influxDBFieldGetNodeInfoNode: param.Node,
		influxDBFieldGetNodeInfoType: param.Type,
		influxDBFieldGetNodeInfoTag:  param.Tag,

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
			influxDBTagGetNodeInfoNode:               param.Node,
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
			influxDBTagGetNodeInfoNode: param.Node,
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
