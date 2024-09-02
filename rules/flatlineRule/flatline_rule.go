package flatlinerule

import (
	"elastalert-go/util"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type FlatlineRule struct {
	Name          string        `yaml:"name"`
	Threshold     int           `yaml:"threshold"`
	Type  				string       			`yaml:"type"`
	Timeframe     time.Duration `yaml:"timeframe"`
	UseCountQuery bool          `yaml:"use_count_query"`
	DocType       string        `yaml:"doc_type"`
	UseTermsQuery bool          `yaml:"use_terms_query"`
	TermsSize     int           `yaml:"terms_size"`
	QueryKey      string        `yaml:"query_key"`
	ForgetKeys    bool          `yaml:"forget_keys"`
	Index         string        `yaml:"index"`
	   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
}

func NewFlatlineRule(name string, threshold int, timeframe time.Duration, useCountQuery bool, docType string, useTermsQuery bool, termsSize int, queryKey string, forgetKeys bool, index string) *FlatlineRule {
	return &FlatlineRule{
		Name:          name,
		Threshold:     threshold,
		Timeframe:     timeframe,
		UseCountQuery: useCountQuery,
		DocType:       docType,
		UseTermsQuery: useTermsQuery,
		TermsSize:     termsSize,
		QueryKey:      queryKey,
		ForgetKeys:    forgetKeys,
		Index:         index,
	}
}

// Matches checks if the total number of events is under the threshold for the timeframe.
func (r *FlatlineRule) Matches(events []map[string]interface{}) bool {
	startTime := time.Now().Add(-r.Timeframe)

	// Count events within the timeframe
	eventCount := 0
	for _, event := range events {
		eventTime, err := time.Parse(time.RFC3339, event["@timestamp"].(string))
		if err != nil {
			fmt.Printf("Error parsing event timestamp: %v\n", err)
			continue
		}
		if eventTime.After(startTime) {
			eventCount++
		}
	}

	// Compare event count with the threshold
	return eventCount < r.Threshold
}

// GetQuery constructs and returns the OpenSearch query for the FlatlineRule.
func (r *FlatlineRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
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
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return &opensearchapi.SearchRequest{}, err
	}

	return &opensearchapi.SearchRequest{
		Index: []string{r.Index},
		Body:  strings.NewReader(string(queryBytes)),
	}, nil
}

// Evaluate processes the query results.
func (r *FlatlineRule) Evaluate(response *opensearchapi.Response) bool {
	hits,_:=util.GetHitsFromResponse(response)
	return r.Matches(hits)
}

func (r *FlatlineRule) GetName() string {
	return r.Name
}

func (r *FlatlineRule) GetIndex() string {
	return r.Index
}
func (r *FlatlineRule) GetType() string {
	return r.Type
}
func (c *FlatlineRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *FlatlineRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}