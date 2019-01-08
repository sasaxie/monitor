package storage

import (
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/javatron/parser"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

const (
	influxDBPointNameChainParameters = "api_chain_parameters"
)

func GetChainParametersStorage(
	db *influxdb.InfluxDB,
	data interface{},
	nodeHost, nodeTagName, nodeType string) error {
	chainParameters, ok := data.(parser.ChainParameters)

	logs.Debug("GetChainParametersStorage storing")

	if !ok {
		return errors.New("GetChainParametersStorage convert error")
	}

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	for key, value := range chainParameters {
		tags[key] = key
		fields[key] = value
	}

	return db.Write(
		influxDBPointNameChainParameters,
		tags,
		fields,
		time.Now())
}
