package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"strings"
	"sync"
)

type NodeController struct {
	beego.Controller
}

type NodesTaskPool map[string]bool

type NodesResult map[string]bool

var m sync.Mutex

// @Title Get all nodes
// @Description get all nodes
// @router /nodes/tag/:tag [get,post]
func (n *NodeController) Nodes() {
	tag := n.GetString(":tag")

	nodesTaskPool := make(NodesTaskPool)
	nodesResult := make(NodesResult)

	if tag == "" && len(tag) == 0 {
		n.Data["json"] = "not found tag"
	} else {
		addresses := models.ServersConfig.GetAddressStringByTag(tag)

		for _, address := range addresses {
			if _, ok := nodesTaskPool[address]; !ok {
				nodesTaskPool[address] = false
			}
		}

		getAllNodes(nodesTaskPool, nodesResult)

		res := make([]string, 0)
		for k := range nodesResult {
			res = append(res, strings.Split(k, ":")[0])
		}
		n.Data["json"] = res
	}

	n.ServeJSON()
}

func getAllNodes(nodesTaskPool NodesTaskPool,
	nodesResult NodesResult) {

	for {
		var wg sync.WaitGroup
		for address, isFinished := range nodesTaskPool {
			if !isFinished {
				wg.Add(1)
				go getAllNodesByAddress(address, &wg, nodesTaskPool, nodesResult)
				nodesTaskPool[address] = true
			}
		}

		wg.Wait()

		finishedCount := 0
		for _, v := range nodesTaskPool {
			if v {
				finishedCount++
			}
		}

		if finishedCount >= len(nodesTaskPool) {
			break
		}
	}
}

func getAllNodesByAddress(address string, wg *sync.WaitGroup,
	nodesTaskPool NodesTaskPool,
	nodesResult NodesResult) []string {

	defer wg.Done()

	var res []string

	client := service.NewGrpcClient(address)
	client.Start()
	defer client.Conn.Close()

	if client != nil {
		res = getNodes(client)
	}

	for _, r := range res {
		m.Lock()
		nodesResult[r] = false

		if _, ok := nodesTaskPool[r]; !ok {
			nodesTaskPool[r] = false
		}
		m.Unlock()
	}

	return res
}

func getNodes(client *service.GrpcClient) []string {
	res := make([]string, 0)

	nodes := client.ListNodes()

	if nodes != nil {
		for _, node := range nodes.Nodes {
			a := fmt.Sprintf("%s:50051", node.Address.Host)

			res = append(res, a)
		}
	}

	return res
}
