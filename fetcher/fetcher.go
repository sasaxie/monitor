package fetcher

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
)

func NilFetcher(url string) ([]byte, error) {
	logs.Debug("nil fetching")
	return nil, nil
}

func DefaultFetcher(url string) ([]byte, error) {
	logs.Debug("default fetching")
	response, err := http.Get(url)

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

	logs.Debug(string(data))

	return data, nil
}
