package frequencyrule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
    

	"github.com/opensearch-project/opensearch-go/opensearchapi"
    util "elastalert-go/util"
)

type FrequencyRule struct {
	Name                  string                  `yaml:"name"`
	Index                 string                  `yaml:"index"`
    Type  				string       			`yaml:"type"`
	NumEvents             int                     `yaml:"num_events"`
	Timeframe             Timeframe               `yaml:"timeframe"`
	TimestampField        string                  `yaml:"timestamp_field"`
	AttachRelated         bool                    `yaml:"attached_related"`
	Priority              int                     `yaml:"priority"`
	Occurrences           map[string]*EventWindow `yaml:"occurrences"`
	Filter        []interface{}             `yaml:"filter"`
   Alert              []string `yaml:"alert"`
    SlackWebhookURL    string   `yaml:"slack_webhook_url"`
	SlackChannelOverride  string                  `yaml:"slack_channel_override"`
	SlackUsernameOverride string                  `yaml:"slack_username_override"`
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


type EventWindow struct {
	Data      []Event
	Timeframe time.Duration
	GetTs     func(event Event) time.Time
}

type Event struct {
	Timestamp     time.Time
	Count         int
	RelatedEvents []Event
}

func NewFrequencyRule(name string, index string, numEvents int, timeframe Timeframe, timestampField string, attachRelated bool, priority int, filter []interface{}, alert []string, slackWebhookURL string, slackChannelOverride string, slackUsernameOverride string) *FrequencyRule {
	return &FrequencyRule{
		Name:                  name,
		Index:                 index,
		NumEvents:             numEvents,
		Timeframe:             timeframe,
		TimestampField:        timestampField,
		AttachRelated:         attachRelated,
		Priority:              priority,
		Occurrences:           make(map[string]*EventWindow),
		Filter:                filter,
		Alert:                 alert,
		SlackWebhookURL:       slackWebhookURL,
		SlackChannelOverride:  slackChannelOverride,
		SlackUsernameOverride: slackUsernameOverride,
	}
}

func (rule *FrequencyRule) AddCountData(data map[time.Time]int) {
	if data == nil {
		fmt.Println("No event data to add")
		return
	}

	if rule.Occurrences == nil {
		fmt.Println("Occurrences map is nil, initializing now.")
		rule.Occurrences = make(map[string]*EventWindow)
	}
	fmt.Println("data is", data)
	for ts, count := range data {
		event := Event{
			Timestamp: ts,
			Count:     count,
		}
		key := "all"
		window, ok := rule.Occurrences[key]
		if !ok {
			window = NewEventWindow(rule.Timeframe.ToDuration())
			rule.Occurrences[key] = window
		}
		window.Append(event)
		rule.Matches(key)
	}
}




func NewEventWindow(timeframe time.Duration) *EventWindow {
	return &EventWindow{
		Data:      make([]Event, 0),
		Timeframe: timeframe,
		GetTs:     func(event Event) time.Time { return event.Timestamp },
	}
}

func (ew *EventWindow) Append(event Event) {
    if ew == nil {
        fmt.Println("EventWindow is nil")
        return
    }
    ew.Data = append(ew.Data, event)
}

func (ew *EventWindow) Count() int {
    if ew == nil || ew.Data == nil {
        return 0
    }
    return len(ew.Data)
}

func (rule *FrequencyRule) Matches(key string) bool {
    if rule == nil || rule.Occurrences == nil {
        fmt.Println("FrequencyRule or Occurrences map is nil")
        return false
    }

    window, ok := rule.Occurrences[key]
    if !ok || window == nil {
        fmt.Printf("Window for key %s is nil\n", key)
        return false
    }

    if window.Count() >= rule.NumEvents {
        lastEvent := window.Data[len(window.Data)-1]
        if rule.AttachRelated {
            relatedEvents := make([]Event, len(window.Data)-1)
            copy(relatedEvents, window.Data[:len(window.Data)-1])
            lastEvent.RelatedEvents = relatedEvents
        }

        fmt.Printf("Match found! %+v\n", lastEvent)

        window.Clear()

        return true
    }

    return false
}


// Clear removes all events from the event window
func (ew *EventWindow) Clear() {
	ew.Data = make([]Event, 0)
}

func (rule *FrequencyRule) GetQuery() (*opensearchapi.SearchRequest, error) {
    // Initialize an empty slice for filters
    var filters []interface{}
	timeframe := ""
    if rule.Timeframe.Minutes > 0 {
        timeframe = fmt.Sprintf("now-%dm", rule.Timeframe.Minutes)
    } else if rule.Timeframe.Hours > 0 {
        timeframe = fmt.Sprintf("now-%dh", rule.Timeframe.Hours)
    } else if rule.Timeframe.Days > 0 {
        timeframe = fmt.Sprintf("now-%dd", rule.Timeframe.Days)
    }
    // Loop over each filter item from the rule's filter configuration
    for _, f := range rule.Filter {
        if f == nil {
            return nil, fmt.Errorf("filter element is nil")
        }

        var queryPart map[string]interface{}
        switch v := f.(type) {
        case map[interface{}]interface{}:
            // Convert to map[string]interface{}
            queryPart = util.ConvertMapKeys(v)
        case map[string]interface{}:
            // Already the correct type
            queryPart = v
        default:
            return nil, fmt.Errorf("unsupported filter type: %T", f)
        }

        // Validate the structure of the queryPart to ensure it matches OpenSearch's expected format
        if _, ok := queryPart["query"]; ok {
            filters = append(filters, queryPart["query"])
        } else {
            return nil, fmt.Errorf("invalid filter format: expected 'query' key, got %v", queryPart)
        }
    }

    // Construct the dynamic query with filters
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "bool": map[string]interface{}{
                "filter": filters, // Use the dynamically created filters
				"must": []interface{}{
                    map[string]interface{}{
                        "range": map[string]interface{}{
                            rule.TimestampField: map[string]interface{}{
                                "gte": timeframe,
                            },
                        },
                    },
                },
            },
        },
    }

    // Serialize the query to JSON
    queryBytes, err := json.Marshal(query)
    if err != nil {
        return &opensearchapi.SearchRequest{}, err
    }

    // Create and return the OpenSearch search request
	// fmt.Printf("Generated Query: %s\n", string(queryBytes))
    return &opensearchapi.SearchRequest{
        Index: []string{rule.Index},
        Body:  strings.NewReader(string(queryBytes)),
    }, nil
}






func (r *FrequencyRule) Evaluate(response *opensearchapi.Response) bool {
    hits,_:=util.GetHitsFromResponse(response)
	if hits == nil {
		fmt.Println("No hits found in the response")
		return false
	}
	fmt.Println("length of hits", len(hits))
	return len(hits) >= r.NumEvents
}










func (r *FrequencyRule) GetName() string {
	return r.Name
}

func (r *FrequencyRule) GetIndex() string {
	return r.Index
}

func (r *FrequencyRule) GetType() string {
	return r.Type
}

func (c *FrequencyRule) GetAlertTypes() []string {
    return c.Alert
}

func (c *FrequencyRule) GetSlackWebhookURL() string {
    return c.SlackWebhookURL
}