package client

import (
	"fmt"
)

func AssertFloat64(value interface{}) (float64, error) {
	switch res := value.(type) {
	case float64:
		return res, nil
	case float32:
		return float64(res), nil
	case int64:
		return float64(res), nil
	case int32:
		return float64(res), nil
	case int16:
		return float64(res), nil
	case int8:
		return float64(res), nil
	case int:
		return float64(res), nil
	case uint64:
		return float64(res), nil
	case uint32:
		return float64(res), nil
	case uint16:
		return float64(res), nil
	case uint8:
		return float64(res), nil
	case uint:
		return float64(res), nil
	case string:
		return 0, fmt.Errorf("string value can not be evaluated")
	default:
		return 0, fmt.Errorf("unknown data type")
	}
}
