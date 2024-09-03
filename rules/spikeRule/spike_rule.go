package spikerule

import (
	"elastalert-go/util"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type SpikeRule struct {
	Name                  string   `yaml:"name"`
	Index                 string   `yaml:"index"`
	Type  				string       			`yaml:"type"`
	SpikeHeight           int      `yaml:"spike_height"`
	SpikeType             string   `yaml:"spike_type"`
	ThresholdCur          int      `yaml:"threshold_cur"`
	Priority              int      `yaml:"priority"`
	Timeframe             Timeframe `yaml:"timeframe"`
	TimestampField        string   `yaml:"timestamp_field"`
	   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
	SlackChannelOverride  string   `yaml:"slack_channel_override"`
	SlackUsernameOverride string   `yaml:"slack_username_override"`
}

type Timeframe struct {
	Minutes int `yaml:"minutes"`
	Hours   int `yaml:"hours"`
	Days    int `yaml:"days"`
}

func NewSpikeRule(name, index string, spikeHeight int, spikeType string, thresholdCur, priority int, timeframe Timeframe, timestampField string, alert []string, slackWebhookURL, slackChannelOverride, slackUsernameOverride string) *SpikeRule {
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

// Updated GetQuery method
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

	// Construct the query with a date histogram and serial diff aggregation
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				rule.TimestampField: map[string]interface{}{
					"gte": timeframe,
					"lt":  "now",
				},
			},
		},
		"aggs": map[string]interface{}{
			"events_over_time": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          rule.TimestampField,
					"fixed_interval": "1h", // Adjust the interval as needed
					"min_doc_count":  0,    // Set to 0 to ensure all intervals are included
				},
				"aggs": map[string]interface{}{
					"doc_count_diff": map[string]interface{}{
						"serial_diff": map[string]interface{}{
							"buckets_path": "_count",
							"lag":          1, // Compare with the previous bucket
						},
					},
					"spike_filter": map[string]interface{}{
						"bucket_selector": map[string]interface{}{
							"buckets_path": map[string]string{
								"docCountDiff": "doc_count_diff",
							},
							"script": fmt.Sprintf("params.docCountDiff != null && params.docCountDiff >= %d", rule.SpikeHeight), // Use spike height threshold
						},
					},
				},
			},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Generated Query: %s\n", string(queryBytes))
	return &opensearchapi.SearchRequest{
		Index: []string{rule.Index},
		Body:  strings.NewReader(string(queryBytes)),
	}, nil
}

func (r *SpikeRule) Evaluate(response *opensearchapi.Response) bool {
	aggregations,_:=util.GetAggregationsFromResponse(response)
	if aggregations == nil {
        return false
    }

    eventsAgg, ok := aggregations["events_over_time"].(map[string]interface{})
    if !ok {
        fmt.Println("Invalid aggregation format")
        return false
    }

    buckets, ok := eventsAgg["buckets"].([]interface{})
    if !ok {
        fmt.Println("Invalid buckets format")
        return false
    }

    for _, bucket := range buckets {
        bucketMap, ok := bucket.(map[string]interface{})
        if !ok {
            continue
        }

        docCount, _ := bucketMap["doc_count"].(float64)
        docCountDiffMap, _ := bucketMap["doc_count_diff"].(map[string]interface{})
        docCountDiff, _ := docCountDiffMap["value"].(float64)

        if r.SpikeType == "up" && docCount >= float64(r.ThresholdCur) && docCountDiff >= float64(r.SpikeHeight) {
            return true
        }
    }

    return false
}




func (r *SpikeRule) GetName() string {
	return r.Name
}

func (r *SpikeRule) GetIndex() string {
	return r.Index
}
func (r *SpikeRule) GetType() string {
	return r.Type
}
func (c *SpikeRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *SpikeRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}