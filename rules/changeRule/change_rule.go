package changerule

import (
	// "context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type ChangeRule struct {
	Name           string                   `yaml:"name"`
	Type           string                   `yaml:"type"`
	Index          string                   `yaml:"index"`
	QueryKey       string                   `yaml:"query_key"`
	CompareKey     string                   `yaml:"compare_key"` // Changed to CompareKey
	IgnoreNull     bool                     `yaml:"ignore_null"`
	Timeframe      Timeframe                `yaml:"timeframe"`
	TsField        string                   `yaml:"timestamp_field"` // Changed to match YAML field
	Occurrences    map[string][]interface{} `yaml:"occurrences"`
	OccurrenceTime map[string]time.Time     `yaml:"occurrence_time"`
	ChangeMap      map[string][]interface{} `yaml:"change_map"`
}

type Timeframe struct {
	Minutes int `yaml:"minutes"`
	Hours   int `yaml:"hours"`
	Days    int `yaml:"days"`
}

func (tf Timeframe) ToDuration() time.Duration {
	return time.Duration(tf.Days)*24*time.Hour +
		time.Duration(tf.Hours)*time.Hour +
		time.Duration(tf.Minutes)*time.Minute
}

func NewChangeRule(name, index, queryKey string, compareKey string, ignoreNull bool, timeframe Timeframe, tsField string) *ChangeRule {
	if queryKey == "" {
		panic("query_key must be specified for ChangeRule")
	}

	return &ChangeRule{
		Name:           name,
		Index:          index,
		QueryKey:       queryKey,
		CompareKey:     compareKey,
		IgnoreNull:     ignoreNull,
		Timeframe:      timeframe,
		TsField:        tsField,
		Occurrences:    make(map[string][]interface{}),
		OccurrenceTime: make(map[string]time.Time),
		ChangeMap:      make(map[string][]interface{}),
	}
}
func (rule *ChangeRule) Matches(event map[string]interface{}) bool {
	fmt.Println("Inside Matches function with event", event)

	if rule.Occurrences == nil {
		rule.Occurrences = make(map[string][]interface{})
	}
	if rule.OccurrenceTime == nil {
		rule.OccurrenceTime = make(map[string]time.Time)
	}
	if rule.ChangeMap == nil {
		rule.ChangeMap = make(map[string][]interface{})
	}

	key := fmt.Sprintf("%v", event[rule.QueryKey])
	currentValue := event[rule.CompareKey]
	eventTimeStr, eventTimeOk := event[rule.TsField].(string)

	fmt.Println("key is", key)
	fmt.Println("current value is", currentValue)
	var eventTimeParsed time.Time
	if eventTimeOk {
		var err error
		eventTimeParsed, err = time.Parse(time.RFC3339, eventTimeStr)
		if err != nil {
			fmt.Println("Error parsing timestamp:", err)
			return false
		}
	}

	if rule.IgnoreNull && currentValue == nil {
		return false
	}

	// Retrieve previous values
	previousValues, exists := rule.Occurrences[key]
	fmt.Println("previousvalues :", previousValues)
	fmt.Println("occurrence is", rule.Occurrences[key])
	if exists {
		fmt.Println("Previous values:", previousValues)

		// Track changes
		changesDetected := false
		for _, previousValue := range previousValues {
			// Ensure timestamp is valid
			lastTime, lastTimeExists := rule.OccurrenceTime[key]
			if lastTimeExists {
				timeDiff := eventTimeParsed.Sub(lastTime)
				if timeDiff <= rule.Timeframe.ToDuration() {
					// Record the change
					rule.ChangeMap[key] = append(rule.ChangeMap[key], []interface{}{previousValue, currentValue})
					fmt.Printf("Change detected: %v -> %v\n", previousValue, currentValue)
					changesDetected = true
				}
			}
		}

		if changesDetected {
			rule.Occurrences[key] = []interface{}{currentValue}
			return true
		}
	}

	// Record the current value and timestamp
	rule.Occurrences[key] = append(rule.Occurrences[key], currentValue)
	if !eventTimeParsed.IsZero() {
		rule.OccurrenceTime[key] = eventTimeParsed
	}

	return false
}

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

func (r *ChangeRule) GetQuery() (*opensearchapi.SearchRequest, error) {
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
							r.TsField: map[string]interface{}{
								"gte": timeframe,
							},
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

	fmt.Println("Constructed query:", string(queryBytes))
	return &opensearchapi.SearchRequest{
		Index: []string{r.Index},
		Body:  strings.NewReader(string(queryBytes)),
	}, nil
}

// // ExecuteQuery performs the OpenSearch query and returns the results
// func (r *ChangeRule) ExecuteQuery(client *opensearch.Client) (bool, error) {
//     // Construct the query
//     req, err := r.GetQuery()
//     if err != nil {
//         return false, err
//     }

//     // Execute the query
//     res, err := req.Do(context.Background(), client)
//     if err != nil {
//         return false, err
//     }
//     defer res.Body.Close()

//     if res.IsError() {
//         return false, fmt.Errorf("error in response: %s", res.String())
//     }

//     // Parse the response
//     var body map[string]interface{}
//     if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
//         return false, err
//     }

//     // Iterate over user buckets to find IP changes
//     users, ok := body["aggregations"].(map[string]interface{})["users"].(map[string]interface{})["buckets"].([]interface{})
//     if !ok {
//         return false, fmt.Errorf("unexpected response format")
//     }

//     for _, user := range users {
//         userMap, ok := user.(map[string]interface{})
//         if !ok {
//             continue
//         }

//         // key := userMap["key"].(string)
//         latestHit, ok := userMap["latest_ip"].(map[string]interface{})["hits"].(map[string]interface{})["hits"].([]interface{})
//         if !ok || len(latestHit) == 0 {
//             continue
//         }

//         // Extract the latest event for this user
//         latestEvent := latestHit[0].(map[string]interface{})["_source"].(map[string]interface{})

//         // Check for changes
//         if r.Matches(latestEvent) {
//             // Record a match
//             r.AddMatch(latestEvent)
//         }
//     }

//     return true, nil
// }

// Evaluate processes the query results.
func (r *ChangeRule) Evaluate(hits []map[string]interface{}) bool {
    changeDetected:=false
	for _, hit := range hits {
		source, ok := hit["_source"].(map[string]interface{})
		if !ok {
			// Log or handle the case where _source is not of the expected type
			fmt.Println("Unexpected _source type in hit:", hit["_source"])
			continue
		}

		if r.Matches(source) {
			changeDetected = true
		}
	}
	return changeDetected
}
