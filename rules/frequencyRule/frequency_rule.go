package frequencyrule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type FrequencyRule struct {
	Name          string                  `yaml:"name"`
	Index         string                  `yaml:"index"`
	NumEvents     int                     `yaml:"num_events"`
	Timeframe     time.Duration           `yaml:"timeframe"`
	TsField       string                  `yaml:"ts_field"`
	AttachRelated bool                    `yaml:"attach_related"`
	Occurrences   map[string]*EventWindow `yaml:"occurrences"`
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

func NewFrequencyRule(name string, index string, numEvents int, timeframe time.Duration, tsField string, attachRelated bool) *FrequencyRule {
	return &FrequencyRule{
		Name:          name,
		Index:         index,
		NumEvents:     numEvents,
		Timeframe:     timeframe,
		TsField:       tsField,
		AttachRelated: attachRelated,
		Occurrences:   make(map[string]*EventWindow),
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
    fmt.Println("data is",data)
	for ts, count := range data {
		event := Event{
			Timestamp: ts,
			Count:     count,
		}
		key := "all"
		window, ok := rule.Occurrences[key]
		if !ok {
			window = NewEventWindow(rule.Timeframe)
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

func (r *FrequencyRule) GetQuery() (*opensearchapi.SearchRequest, error) {
    timeframe := fmt.Sprintf("now-%dh", int(r.Timeframe.Hours()))

    query := map[string]interface{}{
        "query": map[string]interface{}{
            "bool": map[string]interface{}{
                "filter": []interface{}{
                    map[string]interface{}{
                        "term": map[string]interface{}{
                            "isSensitive": true,
                        },
                    },
                    map[string]interface{}{
                        "exists": map[string]interface{}{
                            "field": "action.delete",
                        },
                    },
                    map[string]interface{}{
                        "range": map[string]interface{}{
                            "times.recordedDateTime": map[string]interface{}{
                                "gte": timeframe,
                            },
                        },
                    },
                },
            },
        },
        "sort": []map[string]interface{}{
            {
                "times.recordedDateTime": map[string]interface{}{
                    "order": "asc",
                },
            },
        },
    }

    queryBytes, err := json.Marshal(query)
    if err != nil {
        return &opensearchapi.SearchRequest{}, err
    }

    return &opensearchapi.SearchRequest{
        Index: []string{r.Index},
        Body:  strings.NewReader(string(queryBytes)),
    }, nil
}




// Evaluate processes the query results.
func (r *FrequencyRule) Evaluate(hits []map[string]interface{}) bool {
    if hits == nil {
        fmt.Println("No hits found in the response")
        return false
    }
    fmt.Println("length of hits",len(hits))
    return len(hits)>r.NumEvents
   
}




func (r *FrequencyRule) GetName() string {
	return r.Name
}

func (r *FrequencyRule) GetIndex() string {
	return r.Index
}
