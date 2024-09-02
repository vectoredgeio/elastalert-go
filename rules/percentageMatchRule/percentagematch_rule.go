package percentagematchrule

import (
	"elastalert-go/util"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// PercentageMatchRule defines the structure for the rule
type PercentageMatchRule struct {
	Name                   string                 `yaml:"name"`
	Index                  string                 `yaml:"index"`
	Type  				string       			`yaml:"type"`
	MatchBucketFilter      map[string]interface{} `yaml:"match_bucket_filter"`
	QueryKey               string                 `yaml:"query_key"`
	MinPercentage          float64                `yaml:"min_percentage"`
	MaxPercentage          float64                `yaml:"max_percentage"`
	PercentageFormatString string                 `yaml:"percentage_format_string"`
	   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
}

// NewPercentageMatchRule creates a new instance of PercentageMatchRule
func NewPercentageMatchRule(name, index string, matchBucketFilter map[string]interface{}, queryKey string, minPercentage, maxPercentage float64, percentageFormatString string) *PercentageMatchRule {
	return &PercentageMatchRule{
		Name:                   name,
		Index:                  index,
		MatchBucketFilter:      matchBucketFilter,
		QueryKey:               queryKey,
		MinPercentage:          minPercentage,
		MaxPercentage:          maxPercentage,
		PercentageFormatString: percentageFormatString,
	}
}

// Matches checks if the given percentage is within the defined range
func (r *PercentageMatchRule) Matches(percentage float64) bool {
	if r.MinPercentage > 0 && percentage < r.MinPercentage {
		return true
	}
	if r.MaxPercentage > 0 && percentage > r.MaxPercentage {
		return true
	}
	return false
}

// FormatPercentage formats the percentage according to the specified format
func (r *PercentageMatchRule) FormatPercentage(percentage float64) string {
	if r.PercentageFormatString != "" {
		return fmt.Sprintf(r.PercentageFormatString, percentage)
	}
	return fmt.Sprintf("%.2f%%", percentage)
}

// GetQuery constructs and returns the OpenSearch query for the PercentageMatchRule
func (r *PercentageMatchRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{
						"term": r.MatchBucketFilter,
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"total_documents": map[string]interface{}{
				"value_count": map[string]interface{}{
					"field": r.QueryKey,
				},
			},
			"matching_documents": map[string]interface{}{
				"filter": map[string]interface{}{
					"term": r.MatchBucketFilter,
				},
				"aggs": map[string]interface{}{
					"count": map[string]interface{}{
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

// Evaluate processes the query results to calculate the percentage of matching documents
func (r *PercentageMatchRule) Evaluate(response *opensearchapi.Response) bool {
	responseBody,_:=util.GetHitsFromResponse(response)
	if len(responseBody) == 0 {
		return false
	}

	result := responseBody[0]

	totalDocuments, ok := result["aggregations"].(map[string]interface{})["total_documents"].(map[string]interface{})["value"].(float64)
	if !ok {
		return false
	}

	matchingDocuments, ok := result["aggregations"].(map[string]interface{})["matching_documents"].(map[string]interface{})["count"].(map[string]interface{})["value"].(float64)
	if !ok {
		return false
	}

	percentage := (matchingDocuments / totalDocuments) * 100

	return r.Matches(percentage)
}

// GetName returns the name of the rule
func (r *PercentageMatchRule) GetName() string {
	return r.Name
}

// GetIndex returns the index of the rule
func (r *PercentageMatchRule) GetIndex() string {
	return r.Index
}
func (r *PercentageMatchRule) GetType() string {
	return r.Type
}
func (c *PercentageMatchRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *PercentageMatchRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}
