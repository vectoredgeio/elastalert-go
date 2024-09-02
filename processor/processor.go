package processor

import (
	"elastalert-go/alerts"
	"elastalert-go/config"
	"elastalert-go/queries"
	"elastalert-go/rules"
	"elastalert-go/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func Start(cfg *config.Config) {
	client, err := queries.NewClient(cfg.EsHost, cfg.EsPort, cfg.Username, cfg.Password)
	if err != nil {
		log.Fatalf("Error creating OpenSearch client: %v", err)
	}

	// ACCESS THE FOLDER IN WHICH ALL THE POLICIES ARE DEFINED
	folderPath := "./policies"

	// SLICE TO STORE THE PATH OF ALL YAML FILES DEFINED IN POLICIES FOLDER
	var ruleFiles []string

	dirEntries, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(folderPath, entry.Name())

		// STORE THE YAML FILE'S PATH IN ruleFiles SLICE
		ruleFiles = append(ruleFiles, filePath)

	}

	var loadedRules []rules.Rule

	for _, ruleFile := range ruleFiles {
		rule, err := rules.LoadRule(ruleFile)
		if err != nil {
			log.Printf("Error loading rule from %s: %v", ruleFile, err)
			continue
		}
		loadedRules = append(loadedRules, rule)
	}

	fmt.Println("Loaded rules:", loadedRules)

	// Main loop
	for {
		for _, rule := range loadedRules {
			fmt.Printf("Processing rule: %s (type: %T)\n", rule.GetName(), rule)

			query, err := rule.GetQuery()
			if err != nil {
				log.Printf("Error constructing query for rule %s: %v", rule.GetName(), err)
				continue
			}

			result, err := queries.Query(client, rule.GetIndex(), query, 1000, rule)
			if err != nil {
				log.Printf("Error querying OpenSearch: %v", err)
				continue
			}

			// hits, aggs, err := parseResult(result)
			// fmt.Println("aggs are", aggs)
			if err != nil {
				log.Printf("Error parsing result: %v", err)
				continue
			}

			var triggered bool
			if dualEvalRule, ok := rule.(rules.DualEvaluatable); ok {
				fmt.Println("inside dual evaluatable")
				previousQuery := buildPreviousQuery(query, rule)
				prevResult, err := queries.Query(client, rule.GetIndex(), previousQuery, 1000, rule)
				if err != nil {
					log.Printf("Error querying previous results for rule %s: %v", rule.GetName(), err)
					continue
				}

				previousHits, _, err := parseResult(prevResult)
				if err != nil {
					log.Printf("Error parsing previous result: %v", err)
					continue
				}
				hits,_:=util.GetHitsFromResponse(result)
				triggered = dualEvalRule.EvaluateDual(hits, previousHits)
			} else {
				triggered = rule.Evaluate(result)
			}

			if triggered {
				// var count int
				// if len(hits) > 0 {
				// 	count = len(hits)
				// } else {
				// 	count = len(aggs)
				// }
				message := fmt.Sprintf("Rule %s triggered: %d events found", rule.GetName())

				sendAlerts(rule, message)
			}
		}

		interval, err := time.ParseDuration(cfg.RunEvery)
		if err != nil {
			log.Fatalf("Error parsing run interval: %v", err)
		}
		time.Sleep(interval)
	}
}

func parseResult(result *opensearchapi.Response) ([]map[string]interface{}, map[string]interface{}, error) {
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading result body: %v", err)
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	hits, ok := searchResult["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("unexpected format of hits in search result")
	}

	hitsMap := make([]map[string]interface{}, len(hits))
	for i, hit := range hits {
		hitsMap[i] = hit.(map[string]interface{})
	}

	fmt.Println("search results are", searchResult)

	// Check if "aggregations" exists in the response
	aggsRaw, ok := searchResult["aggregations"]
	if !ok {
		// Aggregations do not exist, return hits and nil for aggregations
		fmt.Println("aggregations key does not exist in the response")
		return hitsMap, nil, nil
	} else {
		fmt.Println("raw aggregations are", aggsRaw)
	}

	aggs, ok := aggsRaw.(map[string]interface{})
	if !ok {
		// Aggregations exist but are not in the expected format
		return hitsMap, nil, fmt.Errorf("unexpected format of aggregations in search result")
	}

	fmt.Println("aggregations are", aggs)

	return hitsMap, aggs, nil
}

func buildPreviousQuery(query *opensearchapi.SearchRequest, rule rules.Rule) *opensearchapi.SearchRequest {
	
	var previousQuery map[string]interface{}
	queryBytes, _ := ioutil.ReadAll(query.Body)
	json.Unmarshal(queryBytes, &previousQuery)

	previousRange := previousQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{})[0].(map[string]interface{})["range"].(map[string]interface{})["@timestamp"].(map[string]interface{})
	previousRange["gte"] = "now-60d" // Example: Adjust this to the required previous window
	previousRange["lte"] = "now-30d" // Example: Adjust this to the required previous window

	previousQueryBytes, _ := json.Marshal(previousQuery)
	return &opensearchapi.SearchRequest{
		Index: query.Index,
		Body:  strings.NewReader(string(previousQueryBytes)),
	}
}

func sendAlerts(rule rules.Rule, message string) {
	fmt.Printf("Send alert triggered by the rule %s", rule.GetName())
	alertTypes := rule.GetAlertTypes()
	alerting := []alerts.Alert{}
	for _, alertType := range alertTypes {
		switch alertType {
		case "slack":
			slackWebhookURL := rule.GetSlackWebhookURL()
			if slackWebhookURL != "" {
				slackAlert := &alerts.SlackAlert{
					WebhookURL: slackWebhookURL,
				}
				alerting = append(alerting, slackAlert)
			} else {
				fmt.Println("Slack webhook URL not provided for rule:", rule.GetName())
			}

		// Add more cases for different alert types
		// case "email":
		//    sendEmailAlert(...)
		// case "google_chat":
		//    sendGoogleChatAlert(...)

		default:
			fmt.Printf("Unknown alert type %s for rule: %s\n", alertType, rule.GetName())
		}
	}
	// Implement actual alerting logic here, e.g., sending emails or Slack messages.
	// emailAlert := &alerts.EmailAlert{
	// 	Recipients: rule.Email,
	// 	SMTPServer: "smtp.gmail.com",
	// 	SMTPPort:   587,
	// 	Username:   "webdevelopers.410@gmail.com",
	// 	Password:   "kiaviikrjehbawrh",
	// }
	// slackAlert := &alerts.SlackAlert{
	// 	WebhookURL: "https://hooks.slack.com/services/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",
	// }
	// googleChatAlert := &alerts.GoogleChatAlert{
	// 	WebhookURL: "https://chat.googleapis.com/v1/spaces/AAAAfkA3ppk/messages?key=AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI&token=pDJwD3JwznvP6v9Fi_E1H18LVN1I-MCeiWl6jaeRTOc",
	// }
	// alerting := []alerts.Alert{slackAlert, googleChatAlert}
	alerts.SendAlerts(alerting, message)
}
