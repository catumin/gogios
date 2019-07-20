package pop3

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

// Client for POP3.
type Client struct {
	Text      *Connection
	conn      net.Conn
	timeout   time.Duration
	useTLS    bool
	tlsConfig *tls.Config
}

// MessageInfo represents the message attributes returned by a LIST command.
type MessageInfo struct {
	Seq  uint32 // Message sequence number
	Size uint32 // Message size in bytes
	UID  string // Message UID
}

type option func(*Client) option

// Noop is a configuration function that does nothing.
func Noop() option {
	return func(c *Client) option {
		return Noop()
	}
}

// UseTLS is a configuration function whose result is passed as a parameter in
// the Dial function. It configures the client to use TLS.
func UseTLS(config *tls.Config) option {
	return func(c *Client) option {
		c.useTLS = true
		c.tlsConfig = config
		return Noop()
	}
}

// UseTimeout is a configuration function whose result is passed as a parameter in
// the Dial function. It configures the client to use timeouts for each POP3 command.
func UseTimeout(timeout time.Duration) option {
	return func(c *Client) option {
		previous := c.timeout
		c.UseTimeouts(timeout)
		return UseTimeout(previous)
	}
}

const (
	protocol      = "tcp"
	lineSeparator = "\n"
)

// Dial connects to the given address and returns a client holding a tcp connection.
// To pass configuration to the Dial function use the methods UseTLS or UseTimeout.
// E.g. c, err = pop3.Dial(address, pop3.UseTLS(tlsConfig), pop3.UseTimeout(timeout))
func Dial(addr string, options ...option) (*Client, error) {
	client := &Client{}
	for _, option := range options {
		option(client)
	}
	var (
		conn net.Conn
		err  error
	)
	if !client.useTLS {
		if client.timeout > time.Duration(0) {
			conn, err = net.DialTimeout(protocol, addr, client.timeout)
		} else {
			conn, err = net.Dial(protocol, addr)
		}
	} else {
		host, _, _ := net.SplitHostPort(addr)
		if client.timeout > time.Duration(0) {
			d := net.Dialer{Timeout: client.timeout}
			conn, err = tls.DialWithDialer(&d, protocol, addr, setServerName(client.tlsConfig, host))
		} else {
			conn, err = tls.Dial(protocol, addr, setServerName(client.tlsConfig, host))

		}
	}
	if err != nil {
		return nil, err
	}
	client.conn = conn
	err = client.initialize()
	if err != nil {
		return nil, err
	}
	return client, nil

}

// NewClient initializeds a client.
// To pass configuration to the NewClient function use the methods UseTLS or UseTimeout.
// E.g. c, err = pop3.Dial(address, pop3.UseTLS(tlsConfig), pop3.UseTimeout(timeout))
func NewClient(conn net.Conn, options ...option) (*Client, error) {
	client := &Client{conn: conn}

	for _, option := range options {
		option(client)
	}

	err := client.initialize()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (client *Client) initialize() (err error) {
	text := NewConnection(client.conn)
	client.Text = text
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.ReadResponse()
	return
}

// UseTimeouts adds a timeout to the client. Timeouts are applied on every
// POP3 command.
func (client *Client) UseTimeouts(timeout time.Duration) {
	client.timeout = timeout
}

// User issues the POP3 User command.
func (client *Client) User(user string) (err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("USER %s", user)
	return
}

// Pass sends the given password to the server. The password is sent
// unencrypted unless the connection is already secured by TLS (via DialTLS or
// some other mechanism).
func (client *Client) Pass(password string) (err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("PASS %s", password)
	return
}

// Auth sends the given username and password to the server.
func (client *Client) Auth(username, password string) (err error) {
	err = client.User(username)
	if err != nil {
		return
	}
	err = client.Pass(password)
	return
}

// Stat retrieves a drop listing for the current maildrop, consisting of the
// number of messages and the total size (in octets) of the maildrop.
// Information provided besides the number of messages and the size of the
// maildrop is ignored. In the event of an error, all returned numeric values
// will be 0.
func (client *Client) Stat() (count, size uint32, err error) {
	client.setDeadline()
	defer client.resetDeadline()
	l, err := client.Text.Cmd("STAT")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(l)
	count, err = stringToUint32(parts[0])
	if err != nil {
		return 0, 0, errors.New("Invalid server response")
	}
	size, err = stringToUint32(parts[1])
	if err != nil {
		return 0, 0, errors.New("Invalid server response")
	}
	return
}

// List returns the size of the message referenced by the sequence number,
// if it exists. If the message does not exist, or another error is encountered,
// the returned size will be 0.
func (client *Client) List(msgSeqNum uint32) (size uint32, err error) {
	client.setDeadline()
	defer client.resetDeadline()
	l, err := client.Text.Cmd("LIST %d", msgSeqNum)
	if err != nil {
		return 0, err
	}
	size, err = stringToUint32(strings.Fields(l)[1])
	if err != nil {
		return 0, errors.New("Invalid server response")
	}
	return size, nil
}

// ListAll returns a list of MessageInfo for all messages, containing their
// sequence number and size.
func (client *Client) ListAll() (msgInfos []*MessageInfo, err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("LIST")
	if err != nil {
		return
	}
	lines, err := client.Text.ReadMultiLines()
	if err != nil {
		return
	}
	msgInfos = make([]*MessageInfo, len(lines))
	for i, line := range lines {
		var seq, size uint32
		fields := strings.Fields(line)
		seq, err = stringToUint32(fields[0])
		if err != nil {
			return
		}
		size, err = stringToUint32(fields[1])
		if err != nil {
			return
		}
		msgInfos[i] = &MessageInfo{
			Seq:  seq,
			Size: size,
		}
	}
	return
}

// Retr downloads and returns the given message. The lines are separated by LF,
// whatever the server sent.
func (client *Client) Retr(msg uint32) (text string, err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("RETR %d", msg)
	if err != nil {
		return "", err
	}
	lines, err := client.Text.ReadMultiLines()
	text = strings.Join(lines, lineSeparator)
	return
}

// Dele marks the given message as deleted.
func (client *Client) Dele(msg uint32) (err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("DELE %d", msg)
	return
}

// Noop does nothing, but will prolong the end of the connection if the server
// has a timeout set.
func (client *Client) Noop() (err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("NOOP")
	return
}

// Rset unmarks any messages marked for deletion previously in this session.
func (client *Client) Rset() (err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("RSET")
	return
}

// Quit sends the QUIT message to the POP3 server and closes the connection.
func (client *Client) Quit() (err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("QUIT")
	if err != nil {
		return err
	}
	client.Text.Close()
	return
}

// UIDl retrieves the unique ID of the message referenced by the sequence number.
func (client *Client) UIDl(msgSeqNum uint32) (uid string, err error) {
	client.setDeadline()
	defer client.resetDeadline()
	line, err := client.Text.Cmd("UIDL %d", msgSeqNum)
	if err != nil {
		return "", err
	}
	uid = strings.Fields(line)[1]
	return
}

// UIDlAll retrieves the unique IDs and sequence number for all messages.
func (client *Client) UIDlAll() (msgInfos []*MessageInfo, err error) {
	client.setDeadline()
	defer client.resetDeadline()
	_, err = client.Text.Cmd("UIDL")
	if err != nil {
		return
	}
	lines, err := client.Text.ReadMultiLines()
	if err != nil {
		return
	}
	msgInfos = make([]*MessageInfo, len(lines))
	for i, line := range lines {
		var seq uint32
		var uid string
		fields := strings.Fields(line)
		seq, err = stringToUint32(fields[0])
		if err != nil {
			return
		}
		uid = fields[1]
		msgInfos[i] = &MessageInfo{
			Seq: seq,
			UID: uid,
		}
	}
	return
}

func (client *Client) setDeadline() {
	if client.timeout > time.Duration(0) {
		client.conn.SetDeadline(time.Now().Add(client.timeout))
	}
}

func (client *Client) resetDeadline() {
	if client.timeout > time.Duration(0) {
		client.conn.SetDeadline(time.Time{})
	}
}

func stringToUint32(intString string) (uint32, error) {
	val, err := strconv.Atoi(intString)
	if err != nil {
		return 0, err
	}
	return uint32(val), nil
}

// setServerName returns a new TLS configuration with ServerName set to host if
// the original configuration was nil or config.ServerName was empty.
// Copied from go-imap: code.google.com/p/go-imap/go1/imap
func setServerName(config *tls.Config, host string) *tls.Config {
	if config == nil {
		config = &tls.Config{ServerName: host}
	} else if config.ServerName == "" {
		c := *config
		c.ServerName = host
		config = &c
	}
	return config
}
