package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

// GetStrValueOfAnyType return string format of any value, for map, need to convert to json
func GetStrValueOfAnyType(value interface{}) string {
	var strVal string
	if _, ok := value.(map[string]interface{}); ok {
		b, err := json.Marshal(value)
		if err != nil {
			log.Fatalf("can not marshal json object, error %v", err)
			return ""
		}
		strVal = string(b)
	} else {
		switch val := value.(type) {
		case float64:
			strVal = strconv.FormatFloat(val, 'f', -1, 64)
		case float32:
			strVal = strconv.FormatFloat(float64(val), 'f', -1, 32)
		default:
			strVal = fmt.Sprintf("%v", value)
		}
	}
	return strVal
}
