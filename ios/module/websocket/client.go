package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/injoyai/tinygo/ios"
	"net/http"
)

var _ ios.MReadWriteCloser = &Client{}

func NewDial(url string) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(url)
		return c, url, err
	}
}

func Dial(url string) (*Client, error) {
	conn, header, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &Client{Conn: conn, Response: header}, nil
}

type Dialer struct {
	websocket.Dialer
	Header http.Header
}

func (this *Dialer) Dial(url string) (*Client, error) {
	conn, header, err := this.Dialer.Dial(url, this.Header)
	if err != nil {
		return nil, err
	}
	return &Client{Conn: conn, Response: header}, nil
}

type Client struct {
	*websocket.Conn
	Response *http.Response
}

func (this *Client) ReadMessage() ([]byte, error) {
	_, bs, err := this.Conn.ReadMessage()
	return bs, err
}

func (this *Client) Write(p []byte) (int, error) {
	//文本传输和二进制传输是一样的,区别在于浏览器是否做UTF-8编码
	err := this.Conn.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}
