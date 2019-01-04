package parser

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
)

type Block struct {
	BlockHeader *BlockHeader `json:"block_header"`
}

type BlockHeader struct {
	RawData *RawData `json:"raw_data"`
}

type RawData struct {
	Number int64 `json:"number"`
}

func NilParser(data []byte) (interface{}, error) {
	logs.Debug("nil parsing")
	return nil, nil
}

func GetNowBlockParser(data []byte) (interface{}, error) {
	logs.Debug("GetNowBlockParser parsing")
	var block Block
	err := json.Unmarshal(data, &block)

	return block, err
}

type WitnessList struct {
	Witnesses []*Witness `json:"witnesses"`
}

type Witness struct {
	Address     string `json:"address"`
	Url         string `json:"url"`
	TotalMissed int64  `json:"totalMissed"`
	IsJobs      bool   `json:"isJobs"`
}

func ListWitnessesParser(data []byte) (interface{}, error) {
	logs.Debug("ListWitnessesParser parsing")
	var witnesses WitnessList
	err := json.Unmarshal(data, &witnesses)

	return witnesses, err
}
