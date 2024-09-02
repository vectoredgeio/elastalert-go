package cardinalityrule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type CardinalityRule struct {
	Name             string            `yaml:"name"`
	Index            string            `yaml:"index"`
	Timeframe        Timeframe     `yaml:"timeframe"`
	CardinalityField string            `yaml:"cardinality_field"`
	MaxCardinality   int               `yaml:"max_cardinality"`
	MinCardinality   int               `yaml:"min_cardinality"`
	QueryKey         string            `yaml:"query_key"`
	Email            []string          `yaml:"email"`
	Occurrences      map[string]int    `yaml:"-"`
	FirstEvent       map[string]time.Time `yaml:"-"`
	Type			string					`yaml:"type"`
	   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
}
type Timeframe struct {
	Minutes int `yaml:"minutes"`
	Hours   int `yaml:"hours"`
	Days    int `yaml:"days"`
}

func (tf Timeframe) ToDuration() time.Duration {
	return time.Duration(tf.Minutes)*time.Minute +
		time.Duration(tf.Hours)*time.Hour +
		time.Duration(tf.Days)*time.Hour*24
}


func NewCardinalityRule(name, index, cardinalityField string, timeframe Timeframe, maxCardinality, minCardinality int, queryKey string, email []string) *CardinalityRule {
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
		if ts.Sub(eventTime) > r.Timeframe.ToDuration() {
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

func (r *CardinalityRule) calculateCardinality(hits []map[string]interface{}) int {
	uniqueValues := make(map[string]struct{})

	for _, hit := range hits {
		fmt.Println("inside loop")
		source, ok := hit["_source"].(map[string]interface{})
		if !ok {
			fmt.Println("_source not found")
			continue
		}
		
		if value, ok := source[r.CardinalityField].(string); ok {
			fmt.Println("accessing username")
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
func (r *CardinalityRule) GetType() string {
	return r.Type
}
func (c *CardinalityRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *CardinalityRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}

// GetQuery constructs and returns the OpenSearch query for the CardinalityRule.
func (r *CardinalityRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	timeframe := ""
    if r.Timeframe.Minutes > 0 {
        timeframe = fmt.Sprintf("now-%dm", r.Timeframe.Minutes)
    } else if r.Timeframe.Hours > 0 {
        timeframe = fmt.Sprintf("now-%dh", r.Timeframe.Hours)
    } else if r.Timeframe.Days > 0 {
        timeframe = fmt.Sprintf("now-%dd", r.Timeframe.Days)
    }
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": timeframe,
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"cardinality_count": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": r.CardinalityField+".keyword",
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



func (r *CardinalityRule) Evaluate(hits []map[string]interface{}) bool {
	cardinality := r.calculateCardinality(hits)

	fmt.Printf("Calculated Cardinality: %d\n", cardinality)
	fmt.Printf("Max Cardinality: %d\n", r.MaxCardinality)
	fmt.Printf("Min Cardinality: %d\n", r.MinCardinality)

	if r.MaxCardinality > 0 {
		if cardinality > r.MaxCardinality {
			fmt.Println("Cardinality exceeds MaxCardinality, returning true")
			return true
		}
	} else if r.MinCardinality > 0 {
		if cardinality < r.MinCardinality {
			fmt.Println("Cardinality is below MinCardinality, returning true")
			return true
		}
	}

	fmt.Println("Cardinality does not meet criteria, returning false")
	return false
}

