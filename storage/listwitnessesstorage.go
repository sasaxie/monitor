package storage

import (
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/javatron/parser"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

const (
	influxDBFieldListWitnessesAddress     = "Address"
	influxDBFieldListWitnessesTotalMissed = "TotalMissed"
	influxDBFieldListWitnessesUrl         = "Url"
	influxDBFieldListWitnessesIsJobs      = "IsJobs"
	influxDBPointNameListWitnesses        = "api_list_witnesses"
)

func ListWitnessesStorage(
	db *influxdb.InfluxDB,
	data interface{},
	nodeHost, nodeTagName, nodeType string) error {
	witnesses, ok := data.(parser.WitnessList)

	logs.Debug("ListWitnessesStorage storing")

	if !ok {
		return errors.New("ListWitnessesStorage convert error")
	}

	saveCount := 0
	if witnesses.Witnesses != nil {
		for _, w := range witnesses.Witnesses {
			if w.IsJobs {
				witnessTags := map[string]string{
					influxDBFieldListWitnessesUrl: w.Url,
				}

				witnessFields := map[string]interface{}{
					influxDBFieldListWitnessesAddress:     w.Address,
					influxDBFieldListWitnessesTotalMissed: w.TotalMissed,
					influxDBFieldListWitnessesUrl:         w.Url,
					influxDBFieldListWitnessesIsJobs:      w.IsJobs,
				}

				err := db.Write(
					influxDBPointNameListWitnesses,
					witnessTags,
					witnessFields,
					time.Now(),
				)

				if err != nil {
					logs.Error(err)
				}
				saveCount++
			}
		}
	}

	logs.Debug("ListWitnessesStorage save", saveCount, "witness")

	return nil
}
