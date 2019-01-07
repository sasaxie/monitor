package storage

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/javatron/parser"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

const (
	influxDBFieldNowBlockNode    = "Node"
	influxDBFieldNowBlockType    = "Type"
	influxDBFieldNowBlockTagName = "TagName"
	influxDBFieldNowBlockNumber  = "Number"
	influxDBPointNameNowBlock    = "api_get_now_block"
)

func GetNowBlockStorage(
	db *influxdb.InfluxDB,
	data interface{},
	nodeHost, nodeTagName, nodeType string) error {
	block, ok := data.(parser.Block)

	logs.Debug("GetNowBlockStorage storing")

	if !ok {
		return errors.New("GetNowBlockStorage convert error")
	}

	nowBlockTags := map[string]string{
		influxDBFieldNowBlockNode:    nodeHost,
		influxDBFieldNowBlockType:    nodeType,
		influxDBFieldNowBlockTagName: nodeTagName,
	}

	nowBlockFields := map[string]interface{}{
		influxDBFieldNowBlockNode:    nodeHost,
		influxDBFieldNowBlockType:    nodeType,
		influxDBFieldNowBlockTagName: nodeTagName,

		influxDBFieldNowBlockNumber: block.BlockHeader.RawData.Number,
	}

	err := db.Write(
		influxDBPointNameNowBlock,
		nowBlockTags,
		nowBlockFields,
		time.Now(),
	)

	if err != nil {
		return err
	}

	logs.Debug(fmt.Sprintf("GetNowBlockStorage save block: #%d",
		block.BlockHeader.RawData.Number))

	return nil
}
