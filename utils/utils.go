package utils

import "encoding/json"

func Dump(data interface{}) string {
	btl, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(btl)
}
