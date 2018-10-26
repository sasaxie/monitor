package influxdb

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/sasaxie/monitor/common/config"
	"log"
	"time"
)

var Client InfluxDB

type InfluxDB struct {
	Addr     string
	Username string
	Password string
	C        client.Client
}

func init() {
	Client = InfluxDB{
		Addr:     config.MonitorConfig.InfluxDB.Address,
		Username: config.MonitorConfig.InfluxDB.Username,
		Password: config.MonitorConfig.InfluxDB.Password,
	}

	var err error
	Client.C, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     Client.Addr,
		Username: Client.Username,
		Password: Client.Password,
	})

	if err != nil {
		log.Fatal(err)
	}

	Client.InitDatabase()
}

func (i *InfluxDB) Write(pointName string, tags map[string]string,
	fields map[string]interface{}) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.MonitorConfig.InfluxDB.Database,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	pt, err := client.NewPoint(pointName, tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	if err := i.C.Write(bp); err != nil {
		log.Fatal(err)
	}
}

func (i *InfluxDB) InitDatabase() {
	_, err := queryDB(i.C, fmt.Sprintf("CREATE DATABASE %s",
		config.MonitorConfig.InfluxDB.Database))
	if err != nil {
		logs.Error(err)
	}
}

func queryDB(clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: config.MonitorConfig.InfluxDB.Database,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
