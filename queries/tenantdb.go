package queries

import (
	// "context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	// "os"
	"time"
	"io/ioutil"
)

// Struct to hold the API response data
type DetectionResponse struct {
	// Define fields based on the response structure
	Data []map[string]interface{} `json:"data"`
	// Other fields can be added as needed
}

// Function to create and send an API request to the tenant detections API
func GetDetections(tenantHost string, tenantPort int, tenantID int, duration string) (*DetectionResponse, error) {
	// Build the URL for the API request
	url := fmt.Sprintf("http://%s:%d/tenant/%d/detections", tenantHost, tenantPort, tenantID)

	// Create a new HTTP client with timeout and TLS config if needed
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Make the GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get detections: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse the JSON response
	var detectionResponse DetectionResponse
	err = json.Unmarshal(body, &detectionResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &detectionResponse, nil
}