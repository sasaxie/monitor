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

	ticker := time.NewTicker(10 * time.Second)
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

		blockNum := 0
		blockID := ""
		if len(nodeInfoDetail.Block) > 0 && !strings.EqualFold(nodeInfoDetail.
			Block, "") {
			strs := strings.Split(nodeInfoDetail.Block, ",")
			if len(strs) > 0 {
				numStrs := strings.Split(strs[0], ":")
				if len(numStrs) > 0 {
					blockNum, err = strconv.Atoi(numStrs[1])
					if err != nil {
						logs.Warn(err)
					}
				}

				idStrs := strings.Split(strs[1], ":")
				if len(idStrs) > 0 {
					blockID = idStrs[1]
				}
			}
		}

		solidityBlockNum := 0
		solidityBlockID := ""
		if len(nodeInfoDetail.SolidityBlock) > 0 && !strings.EqualFold(
			nodeInfoDetail.
				SolidityBlock, "") {
			strs := strings.Split(nodeInfoDetail.SolidityBlock, ",")
			if len(strs) > 0 {
				numStrs := strings.Split(strs[0], ":")
				if len(numStrs) > 0 {
					solidityBlockNum, err = strconv.Atoi(numStrs[1])
					if err != nil {
						logs.Warn(err)
					}
				}

				idStrs := strings.Split(strs[1], ":")
				if len(idStrs) > 0 {
					solidityBlockID = idStrs[1]
				}
			}
		}

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
			hbn := 0
			hbi := ""
			if len(p.HeadBlockWeBothHave) > 0 && !strings.EqualFold(p.
				HeadBlockWeBothHave, "") {
				strs := strings.Split(p.HeadBlockWeBothHave, ",")
				if len(strs) > 0 {
					numStrs := strings.Split(strs[0], ":")
					if len(numStrs) > 0 {
						hbn, err = strconv.Atoi(numStrs[1])
						if err != nil {
							logs.Warn(err)
						}
					}

					idStrs := strings.Split(strs[1], ":")
					if len(idStrs) > 0 {
						hbi = idStrs[1]
					}
				}
			}

			ln := 0
			li := ""
			if len(p.LastSyncBlock) > 0 && !strings.EqualFold(p.
				LastSyncBlock, "") {
				strs := strings.Split(p.LastSyncBlock, ",")
				if len(strs) > 0 {
					numStrs := strings.Split(strs[0], ":")
					if len(numStrs) > 0 {
						ln, err = strconv.Atoi(numStrs[1])
						if err != nil {
							logs.Warn(err)
						}
					}

					idStrs := strings.Split(strs[1], ":")
					if len(idStrs) > 0 {
						li = idStrs[1]
					}
				}
			}

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

			influxdb.Client.Write(config.InfluxDBPointNamePeerInfo, t,
				f)
		}
	}
}
