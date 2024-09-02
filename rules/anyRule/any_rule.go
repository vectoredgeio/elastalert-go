package anyrule

import (
	"elastalert-go/util"
	"encoding/json"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type AnyRule struct {
	Name  string   `yaml:"name"`
	Index string   `yaml:"index"`
	Type string		`yaml:"type"`
	   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
}

func (c *AnyRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *AnyRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}

// NewAnyRule creates a new instance of the AnyRule.
func NewAnyRule(name, index string, email []string) *AnyRule {
	return &AnyRule{
		Name:  name,
		Index: index,
	}
}

// GetQuery constructs and returns the OpenSearch query for the AnyRule.
func (r *AnyRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{}, // Match all documents
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

// Since this is the "Any Rule", it always returns true to indicate an alert should be triggered.
func (r *AnyRule) Evaluate(response *opensearchapi.Response) bool {
	hits,_:=util.GetHitsFromResponse(response)
	// For the "Any Rule", simply return true for any hits received.
	return len(hits) > 0
}

// GetName returns the name of the rule.
func (r *AnyRule) GetName() string {
	return r.Name
}

// GetIndex returns the index associated with the rule.
func (r *AnyRule) GetIndex() string {
	return r.Index
}
func (r *AnyRule) GetType() string {
	return r.Type
}
