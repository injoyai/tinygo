package listen

import (
	"github.com/injoyai/tinygo/ios/module/memory"
	"github.com/injoyai/tinygo/ios/module/mqtt"
	"github.com/injoyai/tinygo/ios/module/tcp"
	"github.com/injoyai/tinygo/ios/module/websocket"
	"github.com/injoyai/tinygo/ios/server"
)

func TCP(port int, op ...server.Option) (*server.Server, error) {
	return server.New(tcp.NewListen(port), op...)
}

func RunTCP(port int, op ...server.Option) error {
	return server.Run(tcp.NewListen(port), op...)
}

func Memory(key string, op ...server.Option) (*server.Server, error) {
	return server.New(memory.NewListen(key), op...)
}

func RunMemory(key string, op ...server.Option) error {
	return server.Run(memory.NewListen(key), op...)
}

func Websocket(port int, op ...server.Option) (*server.Server, error) {
	return server.New(websocket.NewListen(port), op...)
}

func RunWebsocket(port int, op ...server.Option) error {
	return server.Run(websocket.NewListen(port), op...)
}

func MQTT(port int, op ...server.Option) (*server.Server, error) {
	return server.New(mqtt.NewListen(port), op...)
}

func RunMQTT(port int, op ...server.Option) error {
	return server.Run(mqtt.NewListen(port), op...)
}
