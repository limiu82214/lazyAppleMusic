package util

import "encoding/json"

func JsonMarshalWhatever(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
