package ssh

import (
	"context"
	"github.com/injoyai/tinygo/ios"
	"golang.org/x/crypto/ssh"
	"io"
	"strings"
	"time"
)

func NewDial(cfg *Config) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(cfg)
		return c, cfg.Address, err
	}
}

func Dial(cfg *Config) (*Client, error) {
	cfg.init()
	config := &ssh.ClientConfig{
		Timeout:         cfg.Timeout,
		User:            cfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(cfg.Password)},
	}
	switch cfg.Type {
	case "key":
		signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.key), []byte(cfg.keyPassword))
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	sshClient, err := ssh.Dial(cfg.Network, cfg.Address, config)
	if err != nil {
		return nil, err
	}
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	reader, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	outputErr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}
	writer, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := session.RequestPty(cfg.Term, cfg.High, cfg.Wide, cfg.TerminalModes); err != nil {
		return nil, err
	}
	if err := session.Shell(); err != nil {
		return nil, err
	}
	return &Client{
		Writer:  writer,
		Reader:  reader,
		Session: session,
		err:     outputErr,
	}, nil
}

var _ io.ReadWriteCloser = (*Client)(nil)

type Client struct {
	io.Writer
	io.Reader
	*ssh.Session
	err io.Reader
}

type Config struct {
	Address       string
	User          string
	Password      string //类型为password
	Timeout       time.Duration
	High          int               //高
	Wide          int               //宽
	Term          string            //样式 xterm
	Type          string            //password 或者 key
	key           string            //类型为key
	keyPassword   string            //类型为key
	Network       string            //tcp,udp ,可选,默认tcp
	TerminalModes ssh.TerminalModes //可选
}

func (this *Config) init() *Config {
	if !strings.Contains(this.Address, ":") {
		this.Address += ":22"
	}
	if len(this.User) == 0 {
		this.User = "root"
	}
	if this.Timeout == 0 {
		this.Timeout = time.Second
	}
	if this.High == 0 {
		this.High = 32
	}
	if this.Wide == 0 {
		this.Wide = 300
	}
	if len(this.Term) == 0 {
		this.Term = "xterm-256color"
	}
	if len(this.Network) == 0 {
		this.Network = "tcp"
	}
	if this.TerminalModes == nil {
		this.TerminalModes = make(ssh.TerminalModes)
	}
	if this.TerminalModes[ssh.TTY_OP_ISPEED] == 0 {
		this.TerminalModes[ssh.TTY_OP_ISPEED] = 14400 //input speed = 14.4kbaud
	}
	if this.TerminalModes[ssh.TTY_OP_OSPEED] == 0 {
		this.TerminalModes[ssh.TTY_OP_OSPEED] = 14400 //output speed = 14.4kbaud
	}
	if this.TerminalModes[ssh.ECHO] == 0 {
		//禁用回显（0禁用，1启动）
	}
	return this
}
