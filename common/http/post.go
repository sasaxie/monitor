package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func Post(postBody []byte, u string, header map[string]string) ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("POST", u, bytes.NewBuffer(postBody))

	if err != nil {
		return []byte(""), err
	}

	for key, value := range header {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return []byte(""), err
	}

	defer response.Body.Close()

	if err != nil {
		return []byte(""), err
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []byte(""), err
	}

	return data, nil
}
