package cfg

import "io"

var Default = New(nil)

func Init(r io.ReadWriter) {
	Default = New(r)
}

func Set(key string, value any) error {
	return Default.Set(key, value)
}

func Get(key string) any {
	return Default.Get(key).Val()
}

func GetString(key string, def ...string) string {
	return Default.GetString(key, def...)
}

func GetUint8(key string, def ...uint8) uint8 {
	return Default.GetUint8(key, def...)
}

func GetUint16(key string, def ...uint16) uint16 {
	return Default.GetUint16(key, def...)
}

func GetUint32(key string, def ...uint32) uint32 {
	return Default.GetUint32(key, def...)
}

func GetUint64(key string, def ...uint64) uint64 {
	return Default.GetUint64(key, def...)
}

func GetInt8(key string, def ...int8) int8 {
	return Default.GetInt8(key, def...)
}

func GetInt16(key string, def ...int16) int16 {
	return Default.GetInt16(key, def...)
}

func GetInt32(key string, def ...int32) int32 {
	return Default.GetInt32(key, def...)
}

func GetInt64(key string, def ...int64) int64 {
	return Default.GetInt64(key, def...)
}

func GetInt(key string, def ...int) int {
	return Default.GetInt(key, def...)
}

func GetFloat32(key string, def ...float32) float32 {
	return Default.GetFloat32(key, def...)
}

func GetFloat64(key string, def ...float64) float64 {
	return Default.GetFloat64(key, def...)
}

func GetBool(key string, def ...bool) bool {
	return Default.GetBool(key, def...)
}
