package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/common/alarm"
	"github.com/sasaxie/monitor/common/hexutil"
	"github.com/sasaxie/monitor/core"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"strings"
	"sync"
	"time"
)

var TronMonitor *Monitor

func init() {
	TronMonitor = new(Monitor)
	TronMonitor.GRPCMonitor = new(GRPCMonitor)
	TronMonitor.WitnessMonitor = new(WitnessMonitor)
}

type Monitor struct {
	MonitorAddresses []string //此次监控的服务器地址

	GRPCMonitor *GRPCMonitor

	WitnessMonitor *WitnessMonitor
}

func (m *Monitor) Start() {
	beego.Info("start monitor")

	m.GetMonitorAddresses()

	m.GRPCMonitor.Start(m.MonitorAddresses)

	activeAddress := ""
	for k, gRPC := range m.GRPCMonitor.GRPC {
		if gRPC > 0 {
			activeAddress = k
		}
	}

	if !strings.EqualFold(activeAddress, "") {
		m.WitnessMonitor.Start(activeAddress)
	}
}

func (m *Monitor) GetMonitorAddresses() {
	m.MonitorAddresses = models.ServersConfig.GetAllMonitorAddresses()
}

type GRPCMonitor struct {
	LatestGRPCs          map[string][]int64            //最近30次GRPC值
	GRPCMonitorStatusMap map[string]*GRPCMonitorStatus //GRPC监控的数据，用来判断是否需要报警，是否需要提醒恢复
	AlarmMessage         GRPCMonitorMessage            //报警消息
	RecoverMessage       GRPCMonitorMessage            //恢复消息
	GRPC                 map[string]int64              //此次GRPC值
	Mutex                sync.Mutex                    //锁
}

type GRPCMonitorStatus struct {
	Count         int64
	StartPostTime time.Time
}

func (g *GRPCMonitor) UpdateGRPCMonitorStatusMap() {
	if g.GRPCMonitorStatusMap == nil {
		g.GRPCMonitorStatusMap = make(map[string]*GRPCMonitorStatus)
	}

	for address, ping := range g.GRPC {
		if ping <= 0 {
			if v, ok := g.GRPCMonitorStatusMap[address]; ok {
				v.Count = v.Count + 1
				g.Mutex.Lock()
				g.GRPCMonitorStatusMap[address] = v
				g.Mutex.Unlock()
			} else {
				gRPCMonitorStatus := new(GRPCMonitorStatus)
				gRPCMonitorStatus.Count = 1
				g.Mutex.Lock()
				g.GRPCMonitorStatusMap[address] = gRPCMonitorStatus
				g.Mutex.Unlock()
			}
		}
	}
}

func (g *GRPCMonitor) UpdateAlarmMessage() {
	g.AlarmMessage = make(GRPCMonitorMessage)
	// 如果次数>=3，并且时间不足1小时，则发送报警，并重置时间为当前时间
	for k, v := range g.GRPCMonitorStatusMap {
		if (v.Count >= 3) && (time.Now().UTC().Unix()-v.StartPostTime.UTC().
			Unix() >= 3600) {
			g.AlarmMessage[k] = fmt.Sprintf("gRPC接口连续%d次超时(>5000ms)",
				v.Count)
			g.GRPCMonitorStatusMap[k].StartPostTime = time.Now()
		}
	}

	if len(g.AlarmMessage) > 0 {
		bodyContent := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, g.AlarmMessage.String())

		alarm.DingAlarm.Alarm([]byte(bodyContent))
	}
}

func (g *GRPCMonitor) UpdateRecoverMessage() {
	g.RecoverMessage = make(GRPCMonitorMessage)
	// address 没有遍历到的直接移除，如果次数>=3的，还提示恢复信息并从map中移除
	gRPCMonitorStatusMapCopy := make(map[string]*GRPCMonitorStatus)
	for k, v := range g.GRPCMonitorStatusMap {
		gRPCMonitorStatusMapCopy[k] = v
	}

	for k, v := range g.GRPC {
		if v <= 0 {
			delete(gRPCMonitorStatusMapCopy, k)
		}
	}

	for k, v := range gRPCMonitorStatusMapCopy {
		delete(g.GRPCMonitorStatusMap, k)

		if v.Count >= 3 {
			g.RecoverMessage[k] = "gRPC接口已恢复正常"
		}
	}

	if len(g.RecoverMessage) > 0 {
		bodyContent := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, g.RecoverMessage.String())

		alarm.DingAlarm.Alarm([]byte(bodyContent))
	}
}

func (g *GRPCMonitor) UpdateLatestGRPCs() {
	if g.LatestGRPCs == nil {
		g.LatestGRPCs = make(map[string][]int64)
	}

	for address, ping := range g.GRPC {
		if gRPCs, ok := g.LatestGRPCs[address]; ok {
			gRPCs = append(gRPCs, ping)

			if len(gRPCs) > 30 {
				gRPCs = gRPCs[len(gRPCs)-30:]
			}

			g.LatestGRPCs[address] = gRPCs
		} else {
			newGRPCs := make([]int64, 0)
			newGRPCs = append(newGRPCs, ping)
			g.LatestGRPCs[address] = newGRPCs
		}
	}
}

func (g *GRPCMonitor) Start(monitorAddresses []string) {
	g.GRPC = make(map[string]int64)

	var wg sync.WaitGroup

	for _, s := range monitorAddresses {
		wg.Add(1)
		go g.GetGRPC(s, &wg)
	}
	wg.Wait()

	// 进行判断
	// 更新GRPCMonitorStatusMap
	g.UpdateGRPCMonitorStatusMap()

	// 更新恢复消息
	g.UpdateRecoverMessage()

	// 更新报警消息
	g.UpdateAlarmMessage()

	// 更新LatestGRPC
	g.UpdateLatestGRPCs()
}

func (g *GRPCMonitor) GetGRPC(address string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := service.NewGrpcClient(address)
	client.Start()
	defer client.Conn.Close()

	g.Mutex.Lock()
	g.GRPC[address] = client.GetPing()
	g.Mutex.Unlock()
}

type GRPCMonitorMessage map[string]string

func (p GRPCMonitorMessage) String() string {
	message := ""

	for k, v := range p {
		message += fmt.Sprintf("address: %s, message: %s\n", k, v)
	}

	return message
}

type WitnessMonitor struct {
	WitnessMonitorStatusMap map[string]int64 //TotalMissed
}

func (w *WitnessMonitor) Start(activeAddress string) {
	if w.WitnessMonitorStatusMap == nil {
		w.WitnessMonitorStatusMap = make(map[string]int64)
	}

	// 判断Miss Block
	witnessMap := make(map[string]*core.Witness)

	witnesses := service.GrpcClients[activeAddress].ListWitnesses()

	if witnesses != nil {
		for _, witness := range witnesses.Witnesses {
			if witness.IsJobs {
				key := hexutil.Encode(witness.Address)
				if oldTotalMissed, ok := w.WitnessMonitorStatusMap[key]; ok {
					currentTotalMissed := witness.TotalMissed

					if currentTotalMissed > oldTotalMissed {
						witnessMap[key] = witness
					}

					w.WitnessMonitorStatusMap[key] = witness.TotalMissed
				} else {
					w.WitnessMonitorStatusMap[key] = witness.TotalMissed
				}
			}
		}
	}

	if len(witnessMap) > 0 {
		content := ""
		for _, v := range witnessMap {
			content += fmt.Sprintf("[url：%s，当前的totalMissed"+
				"：%d] ", v.Url, v.TotalMissed)
		}

		bodyContent := fmt.Sprintf(`
				{
					"msgtype": "text",
					"text": {
						"content": "超级节点不出块了，一直警告直到恢复正常：%s"
					}
				}
				`, content)

		alarm.DingAlarm.Alarm([]byte(bodyContent))
	}
}

func StartMonitor() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				TronMonitor.Start()
			}
		}
	}()
}
