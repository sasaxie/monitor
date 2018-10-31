package task

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
	"time"
)

func StartHttpMonitor() {
	logs.Info("start http monitor")

	ticker := time.NewTicker(config.MonitorConfig.Task.GetHTTPDataInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, address := range models.NodeList.Addresses {
				go dealHttpMonitor(address.Ip, address.HttpPort)
			}
		}
	}
}

func dealHttpMonitor(ip string, port int) {
	address := fmt.Sprintf("%s:%d", ip, port)
	response, err := http.Get(fmt.Sprintf("http://%s/wallet/getnodeinfo", address))

	if err != nil {
		logs.Debug(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Debug(err)
			return
		}

		var nodeInfoDetail models.NodeInfoDetail
		err = json.Unmarshal(body, &nodeInfoDetail)

		if err != nil {
			logs.Warn(err)
			return
		}

		blockNum, blockID := getBlockNumAndId(nodeInfoDetail.Block)

		solidityBlockNum, solidityBlockID := getBlockNumAndId(nodeInfoDetail.SolidityBlock)

		nodeInfoDetailTags := map[string]string{config.InfluxDBTagNode: address}
		nodeInfoDetailFields := map[string]interface{}{
			config.InfluxDBFieldActiveConnectCount: nodeInfoDetail.
				ActiveConnectCount,
			config.InfluxDBFieldBeginSyncNum: nodeInfoDetail.BeginSyncNum,
			config.InfluxDBFieldBlockNum:     blockNum,
			config.InfluxDBFieldBlockID:      blockID,

			config.InfluxDBFieldActiveNodeSize:           nodeInfoDetail.ConfigNodeInfo.ActiveNodeSize,
			config.InfluxDBFieldAllowCreationOfContracts: nodeInfoDetail.ConfigNodeInfo.AllowCreationOfContracts,
			config.InfluxDBFieldBackupListenPort:         nodeInfoDetail.ConfigNodeInfo.BackupListenPort,
			config.InfluxDBFieldBackupMemberSize:         nodeInfoDetail.ConfigNodeInfo.BackupMemberSize,
			config.InfluxDBFieldBackupPriority:           nodeInfoDetail.ConfigNodeInfo.BackupPriority,
			config.InfluxDBFieldCodeVersion:              nodeInfoDetail.ConfigNodeInfo.CodeVersion,
			config.InfluxDBFieldDbVersion: nodeInfoDetail.
				ConfigNodeInfo.DbVersion,
			config.InfluxDBFieldDiscoverEnable:        nodeInfoDetail.ConfigNodeInfo.DiscoverEnable,
			config.InfluxDBFieldListenPort:            nodeInfoDetail.ConfigNodeInfo.ListenPort,
			config.InfluxDBFieldMaxConnectCount:       nodeInfoDetail.ConfigNodeInfo.MaxConnectCount,
			config.InfluxDBFieldMaxTimeRatio:          nodeInfoDetail.ConfigNodeInfo.MaxTimeRatio,
			config.InfluxDBFieldMinParticipationRate:  nodeInfoDetail.ConfigNodeInfo.MinParticipationRate,
			config.InfluxDBFieldMinTimeRatio:          nodeInfoDetail.ConfigNodeInfo.MinTimeRatio,
			config.InfluxDBFieldP2pVersion:            nodeInfoDetail.ConfigNodeInfo.P2pVersion,
			config.InfluxDBFieldPassiveNodeSize:       nodeInfoDetail.ConfigNodeInfo.PassiveNodeSize,
			config.InfluxDBFieldSameIpMaxConnectCount: nodeInfoDetail.ConfigNodeInfo.SameIpMaxConnectCount,
			config.InfluxDBFieldSendNodeSize:          nodeInfoDetail.ConfigNodeInfo.SendNodeSize,
			config.InfluxDBFieldSupportConstant:       nodeInfoDetail.ConfigNodeInfo.SupportConstant,

			config.InfluxDBFieldCurrentConnectCount: nodeInfoDetail.CurrentConnectCount,

			config.InfluxDBFieldCpuCount: nodeInfoDetail.MachineInfo.CpuCount,
			config.InfluxDBFieldCpuRate: nodeInfoDetail.
				MachineInfo.CpuRate * 100,
			config.InfluxDBFieldDeadLockThreadCount: nodeInfoDetail.MachineInfo.DeadLockThreadCount,
			config.InfluxDBFieldFreeMemory:          nodeInfoDetail.MachineInfo.FreeMemory,
			config.InfluxDBFieldJavaVersion:         nodeInfoDetail.MachineInfo.JavaVersion,
			config.InfluxDBFieldJvmFreeMemory:       nodeInfoDetail.MachineInfo.JvmFreeMemory,
			config.InfluxDBFieldJvmTotalMemoery:     nodeInfoDetail.MachineInfo.JvmTotalMemoery,

			config.InfluxDBFieldOsName: nodeInfoDetail.MachineInfo.OsName,
			config.InfluxDBFieldProcessCpuRate: nodeInfoDetail.MachineInfo.
				ProcessCpuRate * 100,
			config.InfluxDBFieldThreadCount: nodeInfoDetail.MachineInfo.ThreadCount,
			config.InfluxDBFieldTotalMemory: nodeInfoDetail.MachineInfo.TotalMemory,

			config.InfluxDBFieldPassiveConnectCount: nodeInfoDetail.PassiveConnectCount,

			config.InfluxDBFieldSolidityBlockNum: solidityBlockNum,
			config.InfluxDBFieldSolidityBlockID:  solidityBlockID,
			config.InfluxDBFieldTotalFlow:        nodeInfoDetail.TotalFlow,
		}

		influxdb.Client.Write(config.InfluxDBPointNameNodeInfoDetail, nodeInfoDetailTags,
			nodeInfoDetailFields)

		for _, v := range nodeInfoDetail.MachineInfo.MemoryDescInfoList {
			t := map[string]string{config.InfluxDBTagNode: address,
				config.InfluxDBTagMemoryDescInfoName: v.Name}
			f := map[string]interface{}{
				config.InfluxDBFieldMemoryDescInfoInitSize: v.InitSize,
				config.InfluxDBFieldMemoryDescInfoMaxSize:  v.MaxSize,
				config.InfluxDBFieldMemoryDescInfoUseRate:  v.UseRate * 100,
				config.InfluxDBFieldMemoryDescInfoUseSize:  v.UseSize,
			}

			influxdb.Client.Write(config.InfluxDBPointNameNodeInfoDetail, t,
				f)
		}

		for _, p := range nodeInfoDetail.PeerList {
			hbn, hbi := getBlockNumAndId(p.HeadBlockWeBothHave)

			ln, li := getBlockNumAndId(p.LastSyncBlock)

			t := map[string]string{config.InfluxDBTagNode: address,
				config.InfluxDBTagPeer: p.Host}
			f := map[string]interface{}{
				config.InfluxDBFieldPeerActive:                  p.Active,
				config.InfluxDBFieldPeerAvgLatency:              p.AvgLatency,
				config.InfluxDBFieldPeerBlockInPorcSize:         p.BlockInPorcSize,
				config.InfluxDBFieldPeerConnectTime:             p.ConnectTime,
				config.InfluxDBFieldPeerDisconnectTimes:         p.DisconnectTimes,
				config.InfluxDBFieldPeerHeadBlockTimeWeBothHave: p.HeadBlockTimeWeBothHave,
				config.InfluxDBFieldPeerHeadBlockWeBothHaveNum:  hbn,
				config.InfluxDBFieldPeerHeadBlockWeBothHaveID:   hbi,
				config.InfluxDBFieldPeerHost:                    p.Host,
				config.InfluxDBFieldPeerInFlow:                  p.InFlow,
				config.InfluxDBFieldPeerLastBlockUpdateTime:     p.LastBlockUpdateTime,
				config.InfluxDBFieldPeerLastSyncBlockNum:        ln,
				config.InfluxDBFieldPeerLastSyncBlockID:         li,
				config.InfluxDBFieldPeerLocalDisconnectReason:   p.LocalDisconnectReason,
				config.InfluxDBFieldPeerNeedSyncFromPeer:        p.NeedSyncFromPeer,
				config.InfluxDBFieldPeerNeedSyncFromUs:          p.NeedSyncFromUs,
				config.InfluxDBFieldPeerNodeCount:               p.NodeCount,
				config.InfluxDBFieldPeerNodeID:                  p.NodeId,
				config.InfluxDBFieldPeerPort:                    p.Port,
				config.InfluxDBFieldPeerRemainNum:               p.RemainNum,
				config.InfluxDBFieldPeerRemoteDisconnectReason:  p.RemoteDisconnectReason,
				config.InfluxDBFieldPeerScore:                   p.Score,
				config.InfluxDBFieldPeerSyncBlockRequestedSize:  p.SyncBlockRequestedSize,
				config.InfluxDBFieldPeerSyncFlag:                p.SyncFlag,
				config.InfluxDBFieldPeerSyncToFetchSize:         p.SyncToFetchSize,
				config.InfluxDBFieldPeerSyncToFetchSizePeekNum:  p.SyncToFetchSizePeekNum,
				config.InfluxDBFieldPeerUnFetchSynNum:           p.UnFetchSynNum,
			}

			influxdb.Client.Write(config.InfluxDBPointNamePeerInfo, t, f)
		}
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
