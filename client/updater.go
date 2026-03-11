package main

import (
	"encoding/json"
	"fmt"
	log "tools"

	"github.com/go-resty/resty/v2"
)

type updateChecker struct {
	AppName      string
	ServerAddr   string
	localVersion int
}

type Response struct {
	Message string `json:"message"`
	Data    ApplicationRecordWithoutPath
}

type ApplicationRecordWithoutPath struct {
	AppName                     string `json:"appName"`
	AppAvailableVersion         []int  `json:"appAvailableVersion"`
	AppLatestVersion            int    `json:"appLatestVersion"`
	AppForceUpdateMiniumVersion int    `json:"appForceUpdateMiniumVersion"`
	DirectLink                  string `json:"directLink"`
	NoneDirectLink              string `json:"noneDirectLink"`
	Notice                      string `json:"notice"`
}

func (u updateChecker) Check() {
	client := resty.New()
	addr := fmt.Sprintf("http://%s/api/v1/get/%s", u.ServerAddr, u.AppName)
	resp, err := client.R().Get(addr)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("error: Failed to send request to server: %v", err))
	}
	if resp.StatusCode() != 200 {
		log.Logger.Error(fmt.Sprintf("error: Server returned non-200 status code: %d", resp.StatusCode()))
	}
	var record Response
	err = json.Unmarshal(resp.Body(), &record)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("error: Failed to parse server response: %v", err))
	}
	if record.Data.AppLatestVersion > u.localVersion {
		if record.Data.AppForceUpdateMiniumVersion > u.localVersion {
			log.Logger.Warn(fmt.Sprintf("A new version of %s is available (version %d). This update is mandatory. Updating to the latest version.\n", record.Data.AppName, record.Data.AppLatestVersion))
			// place holder update func
		}
	}
}

func main() {
	checker := updateChecker{
		AppName:      "test",
		ServerAddr:   "localhost:8080",
		localVersion: 1,
	}
	checker.Check()

}
