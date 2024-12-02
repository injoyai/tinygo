package codec

import (
	"github.com/injoyai/tinygo/conv/codec/json"
)

type Interface interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

var (
	Default Interface = Json
	Json    Interface = json.Json{}
)
