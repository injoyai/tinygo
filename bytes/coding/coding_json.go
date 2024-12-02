package coding

import "encoding/json"

type Json struct{}

func (this *Json) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (this *Json) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

var (
	JsonMarshal   = json.Marshal
	JsonUnmarshal = json.Unmarshal
)
