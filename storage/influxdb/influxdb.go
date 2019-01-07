package influxdb

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/sasaxie/monitor/common/config"
	"time"
)

type InfluxDB struct {
	Client client.Client
}

func NewInfluxDB(addr, username, password string) (*InfluxDB, error) {
	db := new(InfluxDB)

	var err error
	db.Client, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})

	return db, err
}

func (i *InfluxDB) Write(
	pointName string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time) error {

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.MonitorConfig.InfluxDB.Database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	pt, err := client.NewPoint(pointName, tags, fields, t)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	if err := i.Client.Write(bp); err != nil {
		return err
	}

	return nil
}

func (i *InfluxDB) QueryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: config.MonitorConfig.InfluxDB.Database,
	}
	if response, err := i.Client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}

	return res, nil
}
