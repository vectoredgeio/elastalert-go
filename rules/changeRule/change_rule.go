package changerule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type ChangeRule struct {
	Name               string                   `yaml:"name"`
	Index              string                   `yaml:"index"`
	Type  				string       			`yaml:"type"`
	QueryKey           string                   `yaml:"query_key"`
	CompoundCompareKey []string                 `yaml:"compound_compare_key"`
	IgnoreNull         bool                     `yaml:"ignore_null"`
	Timeframe          time.Duration            `yaml:"timeframe"`
	TsField            string                   `yaml:"ts_field"`
	Occurrences        map[string][]interface{} `yaml:"occurrences"`
	OccurrenceTime     map[string]time.Time     `yaml:"occurrence_time"`
	ChangeMap          map[string][]interface{} `yaml:"change_map"`
}

func NewChangeRule(name, index, queryKey string, compareKeys []string, ignoreNull bool, timeframe time.Duration, tsField string) *ChangeRule {
	if queryKey == "" {
		panic("query_key must be specified for ChangeRule")
	}

	return &ChangeRule{
		Name:               name,
		Index:              index,
		QueryKey:           queryKey,
		CompoundCompareKey: compareKeys,
		IgnoreNull:         ignoreNull,
		Timeframe:          timeframe,
		TsField:            tsField,
		Occurrences:        make(map[string][]interface{}),
		OccurrenceTime:     make(map[string]time.Time),
		ChangeMap:          make(map[string][]interface{}),
	}
}


// Matches checks if values for a certain term change
func (rule *ChangeRule) Matches(event map[string]interface{}) bool {
	key := fmt.Sprintf("%v", event[rule.QueryKey])
	values := make([]interface{}, len(rule.CompoundCompareKey))

	for idx, val := range rule.CompoundCompareKey {
		lookupValue := event[val]
		values[idx] = lookupValue
	}

	changed := false
	for _, val := range values {
		if !rule.IgnoreNull && val == nil {
			return false
		}
	}

	if previousValues, ok := rule.Occurrences[key]; ok {
		for idx, previousValue := range previousValues {
			changed = previousValue != values[idx]
			if changed {
				break
			}
		}

		if changed {
			rule.ChangeMap[key] = []interface{}{previousValues, values}
			if lastTime, ok := rule.OccurrenceTime[key]; ok {
				changed = event[rule.TsField].(time.Time).Sub(lastTime) <= rule.Timeframe
				fmt.Println("value of changes is",changed)
			}
		}
	}

	rule.Occurrences[key] = values
	if _, ok := rule.OccurrenceTime[key]; ok {
		rule.OccurrenceTime[key] = event[rule.TsField].(time.Time)
	}

	return changed
}

// AddMatch adds a match to the rule
func (rule *ChangeRule) AddMatch(match map[string]interface{}) {
	change := rule.ChangeMap[fmt.Sprintf("%v", match[rule.QueryKey])]
	extra := map[string]interface{}{}
	if len(change) > 0 {
		extra = map[string]interface{}{
			"old_value": change[0],
			"new_value": change[1],
		}
	}

	fmt.Println("Match found:", match, extra)
}

func (r *ChangeRule) GetName() string {
	return r.Name
}

func (r *ChangeRule) GetIndex() string {
	return r.Index
}
func (r *ChangeRule) GetType() string {
	return r.Type
}

// GetQuery constructs and returns the OpenSearch query for the ChangeRule.
func (r *ChangeRule) GetQuery() (*opensearchapi.SearchRequest, error) {
	fmt.Println("ts field value is", r.TsField)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							r.TsField: map[string]interface{}{
								"gte": "now-" + r.Timeframe.String(),
							},
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							r.QueryKey: "NotifyFileWasRead", 
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
func (r *ChangeRule) Evaluate(hits []map[string]interface{}) bool {
	for _, hit := range hits {
		source, ok := hit["_source"].(map[string]interface{})
		if !ok {
			// Log or handle the case where _source is not of the expected type
			fmt.Println("Unexpected _source type in hit:", hit["_source"])
			continue
		}

		if r.Matches(source) {
			return true
		}
	}
	return false
}


