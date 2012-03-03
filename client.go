// Copyright (c) 2011 Ross Light.

package ftp

import (
	"errors"
	"io"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
)

// A Client is an FTP client.
type Client struct {
	c       net.Conn
	proto   *textproto.Conn
	Welcome Reply
}

// Dial connects to an FTP server.
func Dial(network, addr string) (*Client, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return NewClient(c)
}

// NewClient creates an FTP client from an existing connection.
func NewClient(c net.Conn) (*Client, error) {
	var err error
	client := &Client{
		c:     c,
		proto: textproto.NewConn(c),
	}
	client.Welcome, err = client.response()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Quit sends the QUIT command and closes the connection.
func (client *Client) Quit() error {
	_, err := client.Do("QUIT")
	if err != nil {
		return err
	}
	return client.Close()
}

// Close closes the connection.
func (client *Client) Close() error {
	return client.proto.Close()
}

// Login sends credentials to the server.
func (client *Client) Login(username, password string) error {
	reply, err := client.Do("USER " + username)
	if err != nil {
		return err
	}
	if reply.Code == CodeNeedPassword {
		reply, err = client.Do("PASS " + password)
		if err != nil {
			return err
		}
	}
	if !reply.PositiveComplete() {
		return reply
	}
	return nil
}

// response reads a reply from the server.
func (client *Client) response() (Reply, error) {
	line, err := client.proto.ReadLine()
	if err != nil {
		return Reply{}, err
	} else if len(line) < 4 {
		return Reply{}, errors.New("Short response line in FTP")
	}

	code, err := strconv.Atoi(line[:3])
	if err != nil {
		return Reply{}, err
	}

	reply := Reply{Code: Code(code)}
	switch line[3] {
	case '-':
		lines := []string{line[4:]}
		endPrefix := strconv.Itoa(code) + " "
		for {
			line, err = client.proto.ReadLine()
			if err != nil {
				break
			}
			if strings.HasPrefix(line, endPrefix) {
				lines = append(lines, line[len(endPrefix):])
				break
			} else {
				lines = append(lines, line)
			}
		}
		reply.Msg = strings.Join(lines, "\n")
		return reply, err
	case ' ':
		reply.Msg = line[4:]
	default:
		return Reply{}, errors.New("Expected space after FTP response code")
	}
	return reply, nil
}

// Do sends a command over the control connection and waits for the response.  It returns any
// protocol error encountered while performing the command.
func (client *Client) Do(command string) (Reply, error) {
	err := client.proto.PrintfLine("%s", command)
	if err != nil {
		return Reply{}, err
	}
	r, err := client.response()
	return r, err
}

// Passive opens a new passive data port.
func (client *Client) Passive() (io.ReadWriteCloser, error) {
	var reply Reply
	var addr *net.TCPAddr
	var err error

	switch client.c.RemoteAddr().Network() {
	case "tcp6":
		reply, err = client.Do("EPSV")
		if err != nil {
			return nil, err
		} else if reply.Code != CodeExtendedPassive {
			return nil, reply
		}

		port, err := parseEpsvReply(reply.Msg)
		if err != nil {
			return nil, err
		}

		addr = &net.TCPAddr{
			IP:   client.c.RemoteAddr().(*net.TCPAddr).IP,
			Port: port,
		}
	default:
		reply, err = client.Do("PASV")
		if err != nil {
			return nil, err
		} else if reply.Code != CodePassive {
			return nil, reply
		}

		addr, err = parsePasvReply(reply.Msg)
		if err != nil {
			return nil, err
		}
	}

	return net.DialTCP("tcp", nil, addr)
}

var pasvRegexp = regexp.MustCompile(`([0-9]+),([0-9]+),([0-9]+),([0-9]+),([0-9]+),([0-9]+)`)

func parsePasvReply(msg string) (*net.TCPAddr, error) {
	numberStrings := pasvRegexp.FindStringSubmatch(msg)
	if numberStrings == nil {
		return nil, errors.New("PASV reply provided no port")
	}
	numbers := make([]byte, len(numberStrings))
	for i, s := range numberStrings {
		n, _ := strconv.Atoi(s)
		numbers[i] = byte(n)
	}
	return &net.TCPAddr{
		IP:   net.IP(numbers[1:5]),
		Port: int(numbers[5])<<8 | int(numbers[6]),
	}, nil
}

const (
	epsvStart = "(|||"
	epsvEnd   = "|)"
)

func parseEpsvReply(msg string) (port int, err error) {
	start := strings.LastIndex(msg, epsvStart)
	if start == -1 {
		return 0, errors.New("EPSV reply provided no port")
	}
	start += len(epsvStart)

	end := strings.LastIndex(msg, epsvEnd)
	if end == -1 || end <= start {
		return 0, errors.New("EPSV reply provided no port")
	}

	return strconv.Atoi(msg[start:end])
}

type transferConn struct {
	io.ReadWriteCloser
	client *Client
}

func (conn transferConn) Close() error {
	err := conn.ReadWriteCloser.Close()
	if err != nil {
		return err
	}
	reply, err := conn.client.response()
	if err != nil {
		return err
	} else if !reply.PositiveComplete() {
		return reply
	}
	return nil
}

// transfer sends a command and opens a new passive data connection.
func (client *Client) transfer(command, dataType string) (io.ReadWriteCloser, error) {
	// Set type
	reply, err := client.Do("TYPE " + dataType)
	switch {
	case err != nil:
		return nil, err
	case !reply.PositiveComplete():
		return nil, reply
	}

	// Open data connection
	conn, err := client.Passive()
	if err != nil {
		return nil, err
	}

	// Send command
	reply, err = client.Do(command)
	switch {
	case err != nil:
		conn.Close()
		return nil, err
	case !reply.Positive():
		conn.Close()
		return nil, reply
	}

	return transferConn{conn, client}, nil
}

// Text sends a command and opens a new passive data connection in ASCII mode.
func (client *Client) Text(command string) (io.ReadWriteCloser, error) {
	return client.transfer(command, "A")
}

// Binary sends a command and opens a new passive data connection in image mode.
func (client *Client) Binary(command string) (io.ReadWriteCloser, error) {
	return client.transfer(command, "I")
}
