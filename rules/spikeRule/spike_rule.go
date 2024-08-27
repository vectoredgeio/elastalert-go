package spikerule

import (
	"encoding/json"
	"fmt"
	"strings"
	
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type SpikeRule struct {
	Name                  string                  `yaml:"name"`
	Index                 string                  `yaml:"index"`
	SpikeHeight           int                     `yaml:"spike_height"`
	SpikeType             string                  `yaml:"spike_type"`
	ThresholdCur          int                     `yaml:"threshold_cur"`
	Priority              int                     `yaml:"priority"`
	Timeframe             Timeframe               `yaml:"timeframe"`
	TimestampField        string                  `yaml:"timestamp_field"`
	Alert                 []string                `yaml:"alert"`
	SlackWebhookURL       string                  `yaml:"slack_webhook_url"`
	SlackChannelOverride  string                  `yaml:"slack_channel_override"`
	SlackUsernameOverride string                  `yaml:"slack_username_override"`
}
type Timeframe struct {
	Minutes int `yaml:"minutes"`
	Hours   int `yaml:"hours"`
	Days    int `yaml:"days"`
}

func NewSpikeRule(name string, index string, spikeHeight int, spikeType string, thresholdCur int, priority int, timeframe Timeframe, timestampField string, alert []string, slackWebhookURL string, slackChannelOverride string, slackUsernameOverride string) *SpikeRule {
	return &SpikeRule{
		Name:                  name,
		Index:                 index,
		SpikeHeight:           spikeHeight,
		SpikeType:             spikeType,
		ThresholdCur:          thresholdCur,
		Priority:              priority,
		Timeframe:             timeframe,
		TimestampField:        timestampField,
		Alert:                 alert,
		SlackWebhookURL:       slackWebhookURL,
		SlackChannelOverride:  slackChannelOverride,
		SlackUsernameOverride: slackUsernameOverride,
	}
}

func (rule *SpikeRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	// Define the timeframe for the query
	timeframe := ""
	if rule.Timeframe.Minutes > 0 {
		timeframe = fmt.Sprintf("now-%dm", rule.Timeframe.Minutes)
	} else if rule.Timeframe.Hours > 0 {
		timeframe = fmt.Sprintf("now-%dh", rule.Timeframe.Hours)
	} else if rule.Timeframe.Days > 0 {
		timeframe = fmt.Sprintf("now-%dd", rule.Timeframe.Days)
	}

	// Construct the query with a date histogram aggregation
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				rule.TimestampField: map[string]interface{}{
					"gte": timeframe,
					"lt":  "now/d",
				},
			},
		},
		"aggs": map[string]interface{}{
			"events_over_time": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          rule.TimestampField,
					"fixed_interval": "1h",
					"min_doc_count":  1,
				},
			},
		},
	}

	// Serialize the query to JSON
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	// Create and return the OpenSearch search request
	fmt.Printf("Generated Query: %s\n", string(queryBytes))
	return &opensearchapi.SearchRequest{
		Index: []string{rule.Index},
		Body:  strings.NewReader(string(queryBytes)),
	}, nil
}

// Evaluate processes the aggregation results to detect spikes.
func (r *SpikeRule) Evaluate(aggregations []map[string]interface{}) bool {
	// Assume aggregations is a slice of maps where each map represents a bucket
	var spikeDetected bool

	for _, aggr := range aggregations {
		eventsAgg, ok := aggr["events_over_time"].(map[string]interface{})
		if !ok {
			fmt.Println("Invalid aggregation format")
			return false
		}

		buckets, ok := eventsAgg["buckets"].([]interface{})
		if !ok {
			fmt.Println("Invalid buckets format")
			return false
		}

		for i, bucket := range buckets {
			bucketMap, ok := bucket.(map[string]interface{})
			if !ok {
				fmt.Println("Invalid bucket format")
				continue
			}

			docCount, _ := bucketMap["doc_count"].(float64)
			if docCount >= float64(r.ThresholdCur) {
				if r.SpikeType == "up" && i > 0 {
					previousBucket := buckets[i-1].(map[string]interface{})
					previousCount, _ := previousBucket["doc_count"].(float64)
					if docCount-previousCount >= float64(r.SpikeHeight) {
						spikeDetected = true
					}
				}
			}
		}
	}

	return spikeDetected
}

func (r *SpikeRule) GetName() string {
	return r.Name
}

func (r *SpikeRule) GetIndex() string {
	return r.Index
}
