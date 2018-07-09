package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func getPrivateIP(apiToken, hostname string) (string, error) {
	url := fmt.Sprintf("https://cloudpanel-api.1and1.com/v1/servers?q=%s&fields=name,private_networks", hostname)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-TOKEN", apiToken)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var servers []struct {
		PrivateNetworks []struct {
			ServerIP string `json:"server_ip"`
		} `json:"private_networks"`
	}
	err = json.NewDecoder(resp.Body).Decode(&servers)
	if err != nil {
		return "", err
	}
	if len(servers) != 1 {
		return "", fmt.Errorf("could not find server for hostname %s", hostname)
	}
	server := servers[0]
	if len(server.PrivateNetworks) != 1 {
		return "", fmt.Errorf("could not find private network for server %s", hostname)
	}
	return server.PrivateNetworks[0].ServerIP, nil
}
