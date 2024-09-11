package queries

import (
	"crypto/tls"
	"elastalert-go/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type APIResponse struct {
	Code   int              `json:"Code"`
	Status bool             `json:"status"`
	Data   []models.Detection `json:"data"`  
}

func GetDetections(tenantHost string, tenantPort int, tenantID int, duration string) ([]models.Detection, error) {
	fmt.Println("get detections called")
	url := fmt.Sprintf("http://%s:%d/tenant/%d/detections", tenantHost, tenantPort, tenantID)

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get detections: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// fmt.Println("body is",string(body.Data))
	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Println("api response is",apiResponse)
	return apiResponse.Data, nil
}
