package folder

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"strings"
)

type Folder struct {
	Id        int64
	Uid       string
	Title     string
	Uri       string
	Url       string
	Type      string
	Tags      []interface{}
	IsStarred bool
}

func GetAllFolders() []*Folder {
	folders := make([]*Folder, 0)

	client := &http.Client{}

	req, err := http.NewRequest("GET",
		"http://127.0.0.1:3000/api/search",
		strings.NewReader("type=dash-folder"))
	if err != nil {
		logs.Warn(err)
		return folders
	}

	req.Header.Set("Authorization", "Bearer eyJrIjoiME5uQ1hLOFljRnNoZVo1UzVtRHVBaHhXcUZocDVCNkoiLCJuIjoiYWRtaW4iLCJpZCI6MX0=")

	resp, err := client.Do(req)
	if err != nil {
		logs.Warn(err)
		return folders
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Warn(err)
		return folders
	}

	err = json.Unmarshal(body, &folders)
	if err != nil {
		logs.Warn(err)
		return folders
	}

	return folders
}
