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

	addresses := m.getWitnessActiveNodes()

	for _, address := range addresses {
		m.WitnessMonitor.Start(address)

	}
}

func (m *Monitor) getWitnessActiveNodes() []string {

	settings := models.ServersConfig.GetSettings()

	addresses := make([]string, 0)
	for _, setting := range settings {
		if strings.EqualFold(setting.IsOpenMonitor, "true") {
			ads := models.ServersConfig.GetAddressStringByTag(setting.Tag)
			for _, ad := range ads {
				if gRPC, ok := m.GRPCMonitor.GRPC[ad]; ok {
					if gRPC > 0 {
						addresses = append(addresses, ad)
						break
					}
				}
			}
		}
	}

	return addresses
}

func (m *Monitor) GetMonitorAddresses() {
	m.MonitorAddresses = models.ServersConfig.GetAllMonitorAddresses()
}

type GRPCMonitor struct {
	LatestGRPCs          map[string]*GRPCs             //最近300次GRPC值
	GRPCMonitorStatusMap map[string]*GRPCMonitorStatus //GRPC监控的数据，用来判断是否需要报警，是否需要提醒恢复
	AlarmMessage         GRPCMonitorMessage            //报警消息
	RecoverMessage       GRPCMonitorMessage            //恢复消息
	GRPC                 map[string]int64              //此次GRPC值
	Mutex                sync.Mutex                    //锁
}

type GRPCs struct {
	Data []int64
	Date []string
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
		g.LatestGRPCs = make(map[string]*GRPCs)
	}

	for address, ping := range g.GRPC {
		if gRPCs, ok := g.LatestGRPCs[address]; ok {
			gRPCs.Data = append(gRPCs.Data, ping)

			gRPCs.Date = append(gRPCs.Date,
				time.Now().Format("2006-01-02 15:04:05"))

			if len(gRPCs.Data) > 300 {
				gRPCs.Data = gRPCs.Data[len(gRPCs.Data)-300:]
				gRPCs.Date = gRPCs.Date[len(gRPCs.Date)-300:]
			}

			g.LatestGRPCs[address] = gRPCs
		} else {
			newGRPCs := new(GRPCs)
			newGRPCs.Data = make([]int64, 0)
			newGRPCs.Date = make([]string, 0)
			newGRPCs.Data = append(newGRPCs.Data, ping)
			newGRPCs.Date = append(newGRPCs.Date,
				time.Now().Format("2006-01-02 15:04:05"))
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
	LatestWitnessMissBlock  map[string]*WitnessMissBlock
}

type WitnessMissBlock struct {
	Data []int64
	Date []string
}

func (w *WitnessMonitor) UpdateLatestWitnessMissBlock(
	key string, totalMissed int64) {

	if w.LatestWitnessMissBlock == nil {
		w.LatestWitnessMissBlock = make(map[string]*WitnessMissBlock)
	}

	if latestWitnessMissBlock, ok := w.LatestWitnessMissBlock[key]; ok {
		latestWitnessMissBlock.Data = append(latestWitnessMissBlock.Data, totalMissed)

		latestWitnessMissBlock.Date = append(latestWitnessMissBlock.Date,
			time.Now().Format("2006-01-02 15:04:05"))

		if len(latestWitnessMissBlock.Data) > 300 {
			latestWitnessMissBlock.Data = latestWitnessMissBlock.Data[len(latestWitnessMissBlock.Data)-300:]
			latestWitnessMissBlock.Date = latestWitnessMissBlock.Date[len(latestWitnessMissBlock.Date)-300:]
		}

		w.LatestWitnessMissBlock[key] = latestWitnessMissBlock
	} else {
		newLatestWitnessMissBlock := new(WitnessMissBlock)
		newLatestWitnessMissBlock.Data = make([]int64, 0)
		newLatestWitnessMissBlock.Date = make([]string, 0)
		newLatestWitnessMissBlock.Data = append(newLatestWitnessMissBlock.Data, totalMissed)
		newLatestWitnessMissBlock.Date = append(newLatestWitnessMissBlock.Date,
			time.Now().Format("2006-01-02 15:04:05"))
		w.LatestWitnessMissBlock[key] = newLatestWitnessMissBlock
	}
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

				w.UpdateLatestWitnessMissBlock(witness.Url, witness.TotalMissed)
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
