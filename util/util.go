package util

import (
	"time"
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
