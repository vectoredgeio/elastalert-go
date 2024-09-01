package metricaggregationrule

import (
    "encoding/json"
    "fmt"
    "math"
    "sort"
    "strings"
    "time"

    "github.com/opensearch-project/opensearch-go/opensearchapi"
)

type MetricAggregationRule struct {
    Name               string            `yaml:"name"`
    Index              string            `yaml:"index"`
    Type  				string       			`yaml:"type"`
    MetricAggKey       string            `yaml:"metric_agg_key"`
    MetricAggType      string            `yaml:"metric_agg_type"`
    MaxThreshold       float64           `yaml:"max_threshold"`
    MinThreshold       float64           `yaml:"min_threshold"`
    PercentileRange    float64           `yaml:"percentile_range"`
    MetricAggScript    string            `yaml:"metric_agg_script"`
    MetricFormatString string            `yaml:"metric_format_string"`
    QueryKey           string            `yaml:"query_key"`
    CompoundQueryKey   []string          `yaml:"compound_query_key"`
   

    CalculationWindow  time.Duration     `yaml:"calculation_window"`
    BufferTime         time.Duration     `yaml:"buffer_time"`
}



func NewMetricAggregationRule(name, index, metricAggKey, metricAggType string, maxThreshold, minThreshold, percentileRange float64, metricAggScript, metricFormatString, queryKey string, compoundQueryKey []string, calculationWindow, bufferTime time.Duration) *MetricAggregationRule {
    return &MetricAggregationRule{
        Name:               name,
        Index:              index,
        MetricAggKey:       metricAggKey,
        MetricAggType:      metricAggType,
        MaxThreshold:       maxThreshold,
        MinThreshold:       minThreshold,
        PercentileRange:    percentileRange,
        MetricAggScript:    metricAggScript,
        MetricFormatString: metricFormatString,
        QueryKey:           queryKey,
        CompoundQueryKey:   compoundQueryKey,
      
        CalculationWindow:  calculationWindow,
        BufferTime:         bufferTime,
    }
}

func (r *MetricAggregationRule) Evaluate(hits []map[string]interface{}) (bool) {
    endTime := time.Now().Add(-r.BufferTime)
    startTime := endTime.Add(-r.CalculationWindow)

    fmt.Printf("Aggregating data between %s and %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

    filteredData := r.filterDataByTime(hits, startTime, endTime)
    metricValue, err := r.extractMetricValue(filteredData)
    if err != nil {
        return false
    }

    return r.checkThresholds(metricValue)
}

func (r *MetricAggregationRule) filterDataByTime(data []map[string]interface{}, startTime, endTime time.Time) []map[string]interface{} {
    var filteredData []map[string]interface{}
    for _, item := range data {
        if timestamp, ok := item["@timestamp"].(time.Time); ok {
            if timestamp.After(startTime) && timestamp.Before(endTime) {
                filteredData = append(filteredData, item)
            }
        }
    }
    return filteredData
}

func (r *MetricAggregationRule) extractMetricValue(data []map[string]interface{}) (float64, error) {
    switch r.MetricAggType {
    case "min":
        return r.calculateMin(data), nil
    case "max":
        return r.calculateMax(data), nil
    case "avg":
        return r.calculateAvg(data), nil
    case "sum":
        return r.calculateSum(data), nil
    case "cardinality":
        return r.calculateCardinality(data), nil
    case "value_count":
        return r.calculateValueCount(data), nil
    case "percentiles":
        return r.calculatePercentiles(data)
    default:
        return 0, fmt.Errorf("unsupported metric aggregation type: %s", r.MetricAggType)
    }
}

func (r *MetricAggregationRule) calculateMin(data []map[string]interface{}) float64 {
    min := math.MaxFloat64
    for _, item := range data {
        if value, ok := item[r.MetricAggKey].(float64); ok {
            if value < min {
                min = value
            }
        }
    }
    return min
}

func (r *MetricAggregationRule) calculateMax(data []map[string]interface{}) float64 {
    max := -math.MaxFloat64
    for _, item := range data {
        if value, ok := item[r.MetricAggKey].(float64); ok {
            if value > max {
                max = value
            }
        }
    }
    return max
}

func (r *MetricAggregationRule) calculateAvg(data []map[string]interface{}) float64 {
    sum := 0.0
    count := 0.0
    for _, item := range data {
        if value, ok := item[r.MetricAggKey].(float64); ok {
            sum += value
            count++
        }
    }
    if count == 0 {
        return 0
    }
    return sum / count
}

func (r *MetricAggregationRule) calculateSum(data []map[string]interface{}) float64 {
    sum := 0.0
    for _, item := range data {
        if value, ok := item[r.MetricAggKey].(float64); ok {
            sum += value
        }
    }
    return sum
}

func (r *MetricAggregationRule) calculateCardinality(data []map[string]interface{}) float64 {
    uniqueValues := make(map[float64]struct{})
    for _, item := range data {
        if value, ok := item[r.MetricAggKey].(float64); ok {
            uniqueValues[value] = struct{}{}
        }
    }
    return float64(len(uniqueValues))
}

func (r *MetricAggregationRule) calculateValueCount(data []map[string]interface{}) float64 {
    count := 0
    for _, item := range data {
        if _, ok := item[r.MetricAggKey].(float64); ok {
            count++
        }
    }
    return float64(count)
}

func (r *MetricAggregationRule) calculatePercentiles(data []map[string]interface{}) (float64, error) {
    values := make([]float64, len(data))
    for i, item := range data {
        if value, ok := item[r.MetricAggKey].(float64); ok {
            values[i] = value
        }
    }
    sort.Float64s(values)
    if len(values) == 0 {
        return 0, fmt.Errorf("no values to calculate percentiles")
    }

    index := int(float64(len(values)-1) * (r.PercentileRange / 100.0))
    return values[index], nil
}

func (r *MetricAggregationRule) checkThresholds(metricValue float64) bool {
    if metricValue == 0 {
        return false
    }
    if r.MaxThreshold != 0 && metricValue > r.MaxThreshold {
        return true
    }
    if r.MinThreshold != 0 && metricValue < r.MinThreshold {
        return true
    }
    return false
}

func (r *MetricAggregationRule) ParseAndEvaluate(responseBody []byte) (bool) {
    var searchResult map[string]interface{}
    if err := json.Unmarshal(responseBody, &searchResult); err != nil {
        return false
    }

    hits, ok := searchResult["hits"].(map[string]interface{})
    if !ok {
        return false
    }

    hitsArray, ok := hits["hits"].([]interface{})
    if !ok {
        return false
    }

    var hitMaps []map[string]interface{}
    for _, hit := range hitsArray {
        hitMap, ok := hit.(map[string]interface{})
        if !ok {
            return false
        }
        hitMaps = append(hitMaps, hitMap)
    }

    return r.Evaluate(hitMaps)
}

// GetQuery constructs and returns the OpenSearch query for the MetricAggregationRule.
func (r *MetricAggregationRule) GetQuery() (*opensearchapi.SearchRequest, error) {
    aggs := make(map[string]interface{})
    switch r.MetricAggType {
    case "min", "max", "avg", "sum", "value_count", "cardinality":
        aggs["metric"] = map[string]interface{}{
            r.MetricAggType: map[string]interface{}{
                "field": r.MetricAggKey,
            },
        }
    case "percentiles":
        aggs["metric"] = map[string]interface{}{
            "percentiles": map[string]interface{}{
                "field":    r.MetricAggKey,
                "percents": []float64{r.PercentileRange},
            },
        }
    default:
        return nil, fmt.Errorf("unsupported metric aggregation type: %s", r.MetricAggType)
    }

    query := map[string]interface{}{
        "query": map[string]interface{}{
            "range": map[string]interface{}{
                "@timestamp": map[string]interface{}{
                    "gte": "now-" + r.CalculationWindow.String(),
                    "lte": "now",
                },
            },
        },
        "aggs": aggs,
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


func (r *MetricAggregationRule) GetName() string {
	return r.Name
}

func (r *MetricAggregationRule) GetIndex() string {
	return r.Index
}
func (r *MetricAggregationRule) GetType() string {
	return r.Type
}