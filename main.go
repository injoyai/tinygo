package main

// This is the most minimal blinky example and should run almost everywhere.

import (
	"github.com/injoyai/tinygo/os/led"
	"machine"
	"time"
)

func main() {

	go led.New(machine.LED).Blink(.5, .5)

	var (
		uart = machine.Serial
		tx   = machine.UART_TX_PIN
		rx   = machine.UART_RX_PIN
	)

	uart.Configure(machine.UARTConfig{BaudRate: 9600, TX: tx, RX: rx})
	uart.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))

	for {
		<-time.After(time.Second)
		uart.Write([]byte("Echo console enabled. Type something then press enter:\r\n"))
	}

}
