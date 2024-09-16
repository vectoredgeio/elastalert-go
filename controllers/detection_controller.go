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

		// Call the methods for each detection
		GetAES(detection)
		GetCES(detection)
	}
	
}

func GetAES(detection models.Detection) {
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
	fmt.Println("detection score is",detectionScore)
	fmt.Printf("Successfully sent detection %s to AES calculation\n", detection.DetectionName)
}


func GetCES(detection models.Detection) {
	// Implement CES logic here
	fmt.Printf("Running GetCES for detection %s\n", detection.DetectionName)
}
