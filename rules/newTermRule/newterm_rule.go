package newtermrule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type NewTermRule struct {
	Name                string            `yaml:"name"`
	Index               string            `yaml:"index"`
	Type  				string       			`yaml:"type"`
	Fields              []interface{}     `yaml:"fields"` // Can be a list of fields or composite fields
	QueryKey            string            `yaml:"query_key"` // If fields is not set, query_key will be used
	TermsWindowSize     time.Duration     `yaml:"terms_window_size"` // Default is 30 days
	WindowStepSize      time.Duration     `yaml:"window_step_size"` // Default is 1 day
	AlertOnMissingField bool              `yaml:"alert_on_missing_field"` // Default is false
	UseTermsQuery       bool              `yaml:"use_terms_query"` // Default is false
	UseKeywordPostfix   bool              `yaml:"use_keyword_postfix"` // Default is true
	Alert               []string          `yaml:"alert"`
	Email               []string          `yaml:"email"`
}

func NewNewTermRule(name, index string, fields []interface{}, queryKey string, alert, email []string) *NewTermRule {
	return &NewTermRule{
		Name:                name,
		Index:               index,
		Fields:              fields,
		QueryKey:            queryKey,
		TermsWindowSize:     30 * 24 * time.Hour, // Default is 30 days
		WindowStepSize:      24 * time.Hour,      // Default is 1 day
		AlertOnMissingField: false,               // Default is false
		UseTermsQuery:       false,               // Default is false
		UseKeywordPostfix:   true,                // Default is true
		Alert:               alert,
		Email:               email,
	}
}

func (r *NewTermRule) GetName() string {
	return r.Name
}

func (r *NewTermRule) GetIndex() string {
	return r.Index
}
func (r *NewTermRule) GetType() string {
	return r.Type
}

// GetQuery constructs the query for the NewTermRule.
func (r *NewTermRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	var fields []string
	for _, field := range r.Fields {
		if r.UseKeywordPostfix {
			fields = append(fields, fmt.Sprintf("%s.keyword", field))
		} else {
			fields = append(fields, fmt.Sprintf("%s", field))
		}
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte": "now-" + r.TermsWindowSize.String(),
								"lte": "now",
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"terms_agg": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": fields[0], 
					"size":  10000,
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



func (r *NewTermRule) Evaluate(hits []map[string]interface{}) bool {
	// This method can remain empty or just return false, as it's not used for NewTermRule.
	return false
}

// Evaluate processes the query results to detect new terms.
func (r *NewTermRule) EvaluateDual(currentResponseBody, previousResponseBody []byte) (bool, error) {
    var currentResult, previousResult map[string]interface{}
    if err := json.Unmarshal(currentResponseBody, &currentResult); err != nil {
        return false, fmt.Errorf("error unmarshalling current response: %v", err)
    }
    if err := json.Unmarshal(previousResponseBody, &previousResult); err != nil {
        return false, fmt.Errorf("error unmarshalling previous response: %v", err)
    }

    // Check if currentResult is nil
    if currentResult == nil {
        return false, fmt.Errorf("current response is nil")
    }

    // Check if previousResult is nil
    if previousResult == nil {
        return false, fmt.Errorf("previous response is nil")
    }

    // Extract current terms
    currentTermsAgg, ok := currentResult["aggregations"].(map[string]interface{})["terms_agg"].(map[string]interface{})
    if !ok {
        return false, fmt.Errorf("unexpected format of current terms aggregation")
    }
    currentBuckets, ok := currentTermsAgg["buckets"].([]interface{})
    if !ok {
        return false, fmt.Errorf("unexpected format of current terms buckets")
    }

    // Extract previous terms
    previousTermsAgg, ok := previousResult["aggregations"].(map[string]interface{})["terms_agg"].(map[string]interface{})
    if !ok {
        return false, fmt.Errorf("unexpected format of previous terms aggregation")
    }
    previousBuckets, ok := previousTermsAgg["buckets"].([]interface{})
    if !ok {
        return false, fmt.Errorf("unexpected format of previous terms buckets")
    }

    currentTermSet := make(map[string]struct{})
    for _, term := range currentBuckets {
        termMap, ok := term.(map[string]interface{})
        if !ok {
            continue
        }
        termKey, ok := termMap["key"].(string)
        if !ok {
            continue
        }
        currentTermSet[termKey] = struct{}{}
    }

    for _, term := range previousBuckets {
        termMap, ok := term.(map[string]interface{})
        if !ok {
            continue
        }
        termKey, ok := termMap["key"].(string)
        if !ok {
            continue
        }
        if _, found := currentTermSet[termKey]; !found {
            return true, nil
        }
    }

    return false, nil
}

