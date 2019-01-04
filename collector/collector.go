package collector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Collector interface {
	Collect()
}

type Common struct {
	Nodes []*Node

	HasInit bool
}

type Node struct {
	CollectionUrl string

	// InfluxDB tags
	Node    string
	Type    string
	TagName string
}

func fetch(u string) ([]byte, error) {
	response, err := http.Get(u)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("fetch error, "+
			"status code: %d", response.StatusCode))
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
