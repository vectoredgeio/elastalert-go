package anyrule

import (
	"encoding/json"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type AnyRule struct {
	Name  string   `yaml:"name"`
	Index string   `yaml:"index"`
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
func (r *AnyRule) Evaluate(hits []map[string]interface{}) bool {
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
