package dashboard

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"strings"
)

type Dashboard struct {
	Id          int64
	Uid         string
	Title       string
	Uri         string
	Url         string
	Type        string
	Tags        []interface{}
	IsStarred   bool
	FolderId    int64
	FolderUid   string
	FolderTitle string
	FolderUrl   string
}

func GetAllDashboards() []*Dashboard {
	dashboards := make([]*Dashboard, 0)

	client := &http.Client{}

	req, err := http.NewRequest("GET",
		"http://127.0.0.1:3000/api/search",
		strings.NewReader("type=dash-db"))
	if err != nil {
		logs.Warn(err)
		return dashboards
	}

	req.Header.Set("Authorization", "Bearer eyJrIjoiME5uQ1hLOFljRnNoZVo1UzVtRHVBaHhXcUZocDVCNkoiLCJuIjoiYWRtaW4iLCJpZCI6MX0=")

	resp, err := client.Do(req)
	if err != nil {
		logs.Warn(err)
		return dashboards
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Warn(err)
		return dashboards
	}

	err = json.Unmarshal(body, &dashboards)
	if err != nil {
		logs.Warn(err)
		return dashboards
	}

	return dashboards
}

func GetDashboardById(id string) *Dashboard {
	dashboard := new(Dashboard)

	client := &http.Client{}

	req, err := http.NewRequest("GET",
		fmt.Sprintf("http://127.0.0.1:3000/api/dashboards/uid/%s", id),
		nil)
	if err != nil {
		logs.Warn(err)
		return dashboard
	}

	req.Header.Set("Authorization", "Bearer eyJrIjoiME5uQ1hLOFljRnNoZVo1UzVtRHVBaHhXcUZocDVCNkoiLCJuIjoiYWRtaW4iLCJpZCI6MX0=")

	resp, err := client.Do(req)
	if err != nil {
		logs.Warn(err)
		return dashboard
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Warn(err)
		return dashboard
	}

	err = json.Unmarshal(body, &dashboard)
	if err != nil {
		logs.Warn(err)
		return dashboard
	}

	return dashboard
}
