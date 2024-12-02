package serial

import (
	"machine"
	"time"
)

func New(uart *machine.UART, tx, rx machine.Pin) *Serial {
	uart.Configure(machine.UARTConfig{BaudRate: 9600, TX: tx, RX: rx})
	return &Serial{
		UART: uart,
	}
}

type Serial struct {
	*machine.UART //串口实例
}

func (this *Serial) SetConfig(config Config) error {
	this.UART.SetBaudRate(config.BaudRate)
	return this.UART.SetFormat(config.Databits, config.Stopbits, config.Parity)
}

// Read 实现io.Reader接口
func (this *Serial) Read(p []byte) (n int, err error) {
	if this.UART.Buffered() == 0 {
		<-time.After(time.Millisecond * 10)
	}
	return this.UART.Read(p)
}

//func (this *Serial) Close() error {
//	return nil
//}

type Parity = machine.UARTParity

type Config struct {
	BaudRate uint32
	Databits uint8
	Stopbits uint8
	Parity   Parity
}

const (
	ParityNone Parity = iota
	ParityOdd
	ParityEven
)
