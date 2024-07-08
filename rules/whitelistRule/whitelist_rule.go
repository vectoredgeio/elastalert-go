package whitelistrule

import (
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type WhitelistRule struct {
	Name       string   `yaml:"name"`
	Index      string   `yaml:"index"`
	CompareKey string   `yaml:"compare_key"`
	Whitelist  []string `yaml:"whitelist"`
	IgnoreNull bool     `yaml:"ignore_null"`
	Email      []string `yaml:"email"`
}

func NewWhitelistRule(name, index, compareKey string, whitelist []string, ignoreNull bool, email []string) *WhitelistRule {
	return &WhitelistRule{
		Name:       name,
		Index:      index,
		CompareKey: compareKey,
		Whitelist:  whitelist,
		IgnoreNull: ignoreNull,
		Email:      email,
	}
}

func (r *WhitelistRule) Matches(event map[string]interface{}) bool {
	value, ok := event[r.CompareKey].(string)
	if !ok {
		return !r.IgnoreNull
	}

	for _, whitelistedValue := range r.Whitelist {
		if value == whitelistedValue {
			return false
		}
	}
	return true
}

// GetQuery constructs and returns the OpenSearch query for the WhitelistRule.
func (r *WhitelistRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	var queryStrings []string

	for _, whitelistedValue := range r.Whitelist {
		queryStrings = append(queryStrings, fmt.Sprintf(`{"term": {"%s": "%s"}}`, r.CompareKey, whitelistedValue))
	}

	queryBody := fmt.Sprintf(`{"query": {"bool": {"must_not": [%s]}}}`, strings.Join(queryStrings, ","))

	return &opensearchapi.SearchRequest{
		Index: []string{r.Index},
		Body:  strings.NewReader(queryBody),
	}, nil
}

// Evaluate processes the query results to determine if any event matches the whitelist criteria.
func (r *WhitelistRule) Evaluate(hits []map[string]interface{}) bool {
	for _, hit := range hits {
		if r.Matches(hit["_source"].(map[string]interface{})) {
			return true
		}
	}
	return false
}
func (r *WhitelistRule) GetName() string {
	return r.Name
}

func (r *WhitelistRule) GetIndex() string {
	return r.Index
}