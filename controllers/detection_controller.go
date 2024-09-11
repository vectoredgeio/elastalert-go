package controllers

import (
	"bytes"
	"elastalert-go/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func ProcessDetections(detections []models.Detection) {
	lastdetection:=detections[len(detections)-4]
	// for _, detection := range detections {
	// 	fmt.Printf("Processing detection: %s\n", detection.DetectionName)

	// 	// Call the methods for each detection
	// 	GetAES(detection)
	// 	GetCES(detection)
	// }
	fmt.Println("detection being sent",lastdetection)
	GetAES(lastdetection)
}

func GetAES(detection models.Detection) {
	// Serialize the detection object to JSON
	detectionJSON, err := json.Marshal(detection)
	if err != nil {
		fmt.Printf("Error marshaling detection data: %v\n", err)
		return
	}
	fmt.Println("jsondetection being sent",detection)
	// Prepare the POST request
	url := "http://0.0.0.0:8089/calculate/AES"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(detectionJSON))
	if err != nil {
		fmt.Printf("Error creating POST request: %v\n", err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the POST request
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

	fmt.Printf("Successfully sent detection %s to AES calculation\n", detection.DetectionName)
}

func GetCES(detection models.Detection) {
	// Implement CES logic here
	fmt.Printf("Running GetCES for detection %s\n", detection.DetectionName)
}
