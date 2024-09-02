package blacklistrule

import (
	"encoding/json"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type BlacklistRule struct {
	Name       string   `yaml:"name"`
	Index      string   `yaml:"index"`
	Type 	string 		`yaml:"type"`
	CompareKey string   `yaml:"compare_key"`
	Blacklist  []string `yaml:"blacklist"`
	Email      []string `yaml:"email"`
	IgnoreNull bool     `yaml:"ignore_null"`
	   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
}

func NewBlacklistRule(name, index, compareKey string, blacklist []string, email []string, ignoreNull bool) *BlacklistRule {
	return &BlacklistRule{
		Name:       name,
		Index:      index,
		CompareKey: compareKey,
		Blacklist:  blacklist,
		Email:      email,
		IgnoreNull: ignoreNull,
	}
}

func (r *BlacklistRule) Matches(event map[string]interface{}) bool {
	value, ok := event[r.CompareKey].(string)
	if !ok {
		return !r.IgnoreNull
	}

	for _, blacklistedValue := range r.Blacklist {
		if strings.Contains(value, blacklistedValue) {
			return true
		}
	}
	return false
}

func (r *BlacklistRule) GetName() string {
	return r.Name
}

func (r *BlacklistRule) GetIndex() string {
	return r.Index
}
func (r *BlacklistRule) GetType() string {
	return r.Type
}
func (c *BlacklistRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *BlacklistRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}
// GetQuery constructs and returns the OpenSearch query for the BlacklistRule.
func (r *BlacklistRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"terms": map[string]interface{}{
							r.CompareKey: r.Blacklist,
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

// Evaluate processes the query results.
func (r *BlacklistRule) Evaluate(hits []map[string]interface{}) bool {
	for _, hit := range hits {
		if r.Matches(hit["_source"].(map[string]interface{})) {
			return true
		}
	}
	return false
}
