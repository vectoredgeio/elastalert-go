package cardinalityrule

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type CardinalityRule struct {
	Name             string            `yaml:"name"`
	Index            string            `yaml:"index"`
	Timeframe        time.Duration     `yaml:"timeframe"`
	CardinalityField string            `yaml:"cardinality_field"`
	MaxCardinality   int               `yaml:"max_cardinality"`
	MinCardinality   int               `yaml:"min_cardinality"`
	QueryKey         string            `yaml:"query_key"`
	Email            []string          `yaml:"email"`
	Occurrences      map[string]int    `yaml:"-"`
	FirstEvent       map[string]time.Time `yaml:"-"`
}

func NewCardinalityRule(name, index, cardinalityField string, timeframe time.Duration, maxCardinality, minCardinality int, queryKey string, email []string) *CardinalityRule {
	return &CardinalityRule{
		Name:             name,
		Index:            index,
		Timeframe:        timeframe,
		CardinalityField: cardinalityField,
		MaxCardinality:   maxCardinality,
		MinCardinality:   minCardinality,
		QueryKey:         queryKey,
		Email:            email,
		Occurrences:      make(map[string]int),
		FirstEvent:       make(map[string]time.Time),
	}
}



func (r *CardinalityRule) GarbageCollect(ts time.Time) {
	for key, eventTime := range r.FirstEvent {
		if ts.Sub(eventTime) > r.Timeframe {
			delete(r.Occurrences, key)
			delete(r.FirstEvent, key)
		}
	}
}

func (r *CardinalityRule) GetKeys() []string {
	keys := make([]string, 0, len(r.Occurrences))
	for key := range r.Occurrences {
		keys = append(keys, key)
	}
	return keys
}

// calculateCardinality calculates the cardinality based on the current hits.
func (r *CardinalityRule) calculateCardinality(hits []map[string]interface{}) int {
	uniqueValues := make(map[string]struct{})

	for _, hit := range hits {
		if value, ok := hit[r.CardinalityField].(string); ok {
			uniqueValues[value] = struct{}{}
		}
	}

	return len(uniqueValues)
}

func (r *CardinalityRule) GetName() string {
	return r.Name
}

func (r *CardinalityRule) GetIndex() string {
	return r.Index
}

// GetQuery constructs and returns the OpenSearch query for the CardinalityRule.
func (r *CardinalityRule) GetQuery() (*opensearchapi.SearchRequest, error) {
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
		"aggs": map[string]interface{}{
			"cardinality_count": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": r.CardinalityField,
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
func (r *CardinalityRule) Evaluate(hits []map[string]interface{}) bool {
	cardinality := r.calculateCardinality(hits)

	if r.MaxCardinality > 0 {
		if cardinality > r.MaxCardinality {
			return true
		}
	} else if r.MinCardinality > 0 {
		if cardinality < r.MinCardinality {
			return true
		}
	}

	return false
}

