package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}
func ConvertMapKeys(m map[interface{}]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    for k, v := range m {
        keyStr, ok := k.(string)
        if !ok {
            continue
        }
        switch v := v.(type) {
        case map[interface{}]interface{}:
            result[keyStr] = ConvertMapKeys(v)
        case []interface{}:
            result[keyStr] = ConvertSlice(v)
        default:
            result[keyStr] = v
        }
    }
    return result
}

func ConvertSlice(s []interface{}) []interface{} {
    result := make([]interface{}, len(s))
    for i, v := range s {
        switch v := v.(type) {
        case map[interface{}]interface{}:
            result[i] = ConvertMapKeys(v)
        case []interface{}:
            result[i] = ConvertSlice(v)
        default:
            result[i] = v
        }
    }
    return result
}
func GetHitsFromResponse(response *opensearchapi.Response) ([]map[string]interface{}, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading result body: %v", err)
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	hits, ok := searchResult["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format of hits in search result")
	}

	hitsMap := make([]map[string]interface{}, len(hits))
	for i, hit := range hits {
		hitsMap[i] = hit.(map[string]interface{})
	}

	return hitsMap, nil
}

func GetAggregationsFromResponse(response *opensearchapi.Response) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading result body: %v", err)
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	aggs, ok := searchResult["aggregations"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected format of aggregations in search result")
	}

	return aggs, nil
}
