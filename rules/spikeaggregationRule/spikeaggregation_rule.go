package spikeaggregationrule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type SpikeAggregationRule struct {
	Name       string        `yaml:"name"`
	Index      string        `yaml:"index"`
	Type  				string       			`yaml:"type"`
	QueryKey   string        `yaml:"query_key"`
	Timeframe  time.Duration `yaml:"timeframe"`
	SpikeHeight float64      `yaml:"spike_height"`
	SpikeType  string        `yaml:"spike_type"`
	Filter     map[string]interface{} `yaml:"filter"`
}

func NewSpikeAggregationRule(name, index, queryKey string, timeframe time.Duration, spikeHeight float64, spikeType string, filter map[string]interface{}) *SpikeAggregationRule {
	return &SpikeAggregationRule{
		Name:       name,
		Index:      index,
		QueryKey:   queryKey,
		Timeframe:  timeframe,
		SpikeHeight: spikeHeight,
		SpikeType:  spikeType,
		Filter:     filter,
	}
}

// GetQuery constructs and returns the OpenSearch query for the SpikeAggregationRule.
func (r *SpikeAggregationRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	spikeWidthStr := fmt.Sprintf("%dh", int(r.Timeframe.Hours()))
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					r.Filter,
					map[string]interface{}{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte": "now-" + r.Timeframe.String(),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"spikes": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field": "@timestamp",
					"interval": spikeWidthStr,
				},
				"aggs": map[string]interface{}{
					"spike_count": map[string]interface{}{
						"value_count": map[string]interface{}{
							"field": r.QueryKey,
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

	return &opensearchapi.SearchRequest{
		Index: []string{r.Index},
		Body:  strings.NewReader(string(queryBytes)),
	}, nil
}

// Evaluate processes the query results to check for spikes.
func (r *SpikeAggregationRule) Evaluate(responseBody []map[string]interface{}) bool {
	if len(responseBody) == 0 {
		return false
	}

	result := responseBody[0]

	aggregations, ok := result["aggregations"].(map[string]interface{})
	if !ok {
		return false
	}

	spikes, ok := aggregations["spikes"].(map[string]interface{})
	if !ok {
		return false
	}

	buckets, ok := spikes["buckets"].([]interface{})
	if !ok {
		return false
	}

	for _, bucket := range buckets {
		bucketMap, ok := bucket.(map[string]interface{})
		if !ok {
			continue
		}

		spikeCount, ok := bucketMap["spike_count"].(map[string]interface{})["value"].(float64)
		if !ok {
			continue
		}

		if r.SpikeType == "up" && spikeCount > r.SpikeHeight {
			return true
		} else if r.SpikeType == "down" && spikeCount < r.SpikeHeight {
			return true
		} else if r.SpikeType == "both" && (spikeCount > r.SpikeHeight || spikeCount < r.SpikeHeight) {
			return true
		}
	}

	return false
}

func (r *SpikeAggregationRule) GetName() string {
	return r.Name
}

func (r *SpikeAggregationRule) GetIndex() string {
	return r.Index
}
func (r *SpikeAggregationRule) GetType() string {
	return r.Type
}
