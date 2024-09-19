package controllers

import (
	"bytes"
	"elastalert-go/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ProcessDetections(detections []models.Detection) {
	for _, detection := range detections {
		fmt.Printf("Processing detection: %s\n", detection.DetectionName)

		CalculateAES(detection)
	}
	
}

func CalculateAES(detection models.Detection) {
	detectionJSON, err := json.Marshal(detection)
	if err != nil {
		fmt.Printf("Error marshaling detection data: %v\n", err)
		return
	}
	url := "http://18.118.97.249:8089/calculate/AES"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(detectionJSON))
	if err != nil {
		fmt.Printf("Error creating POST request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read and handle the response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected response code: %d\n", resp.StatusCode)
		return
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// Print the response
	fmt.Printf("Response from API: %s\n", string(body))
	var aesResponse models.AESResponse

	// Unmarshal the response into the struct
	err = json.Unmarshal(body, &aesResponse)
	if err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return
	}

	// Extract and print the AES score (from the "data" field)
	aesScore := aesResponse.Data

	// var detectionScore models.DetectionScore

	detectionScore:=detection.TransformIntoDetectionScore(detection,aesScore)
	err = SendDetectionScore(detection.TenantID, detectionScore)
	if err != nil {
		fmt.Printf("Error sending detection score: %v\n", err)
		return
	}

	fmt.Printf("Successfully sent detection %s to AES calculation and score endpoint\n", detection.DetectionName)
	
}


func SendDetectionScore(tenantID int, detectionScore models.DetectionScore) error {
	// Marshal detectionScore into JSON

	fmt.Println("detection score query string",detectionScore.QueryString)
	detectionScoreJSON, err := json.Marshal(detectionScore)
	if err != nil {
		return fmt.Errorf("Error marshaling detectionScore: %v", err)
	}

	fmt.Println("detection score json is",detectionScoreJSON)
	url := fmt.Sprintf("http://18.118.97.249:8083/tenant/%d/detections/score", tenantID) 

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(detectionScoreJSON))
	if err != nil {
		return fmt.Errorf("Error creating POST request for detectionScore: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the POST request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected response code when sending detection score: %d", resp.StatusCode)
	}

	// Optionally, read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}
	fmt.Printf("Response from tenant score API: %s\n", string(body))

	return nil
}