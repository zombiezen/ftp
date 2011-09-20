package ftp

import (
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"net"
	"net/textproto"
)

type Client struct {
	c       net.Conn
	proto   *textproto.Conn
	Welcome Reply
}

func Dial(network, addr string) (*Client, os.Error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return NewClient(c)
}

func NewClient(c net.Conn) (*Client, os.Error) {
	var err os.Error
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
func (client *Client) Quit() os.Error {
	_, err := client.Do("QUIT")
	if err != nil {
		return err
	}
	return client.Close()
}

// Close closes the connection.
func (client *Client) Close() os.Error {
	return client.proto.Close()
}

// Login sends credentials to the server.
func (client *Client) Login(username, password string) os.Error {
	reply, err := client.Do("USER " + username)
	if err != nil {
		return err
	}
	if reply.Code >= 300 && reply.Code < 400 {
		reply, err = client.Do("PASS " + password)
	}
	if !reply.Positive() {
		return reply
	}
	return nil
}

// response reads a reply from the server.
func (client *Client) response() (Reply, os.Error) {
	line, err := client.proto.ReadLine()
	if err != nil {
		return Reply{}, err
	} else if len(line) < 4 {
		return Reply{}, os.NewError("Short response line in FTP")
	}

	code, err := strconv.Atoi(line[:3])
	if err != nil {
		return Reply{}, err
	}

	reply := Reply{Code: code}
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
		return Reply{}, os.NewError("Expected space after FTP response code")
	}
	return reply, nil
}

// Do sends a command over the control connection and waits for the response.
// It returns any protocol error encountered while performing the command.
func (client *Client) Do(command string) (Reply, os.Error) {
	err := client.proto.PrintfLine("%s", command)
	if err != nil {
		return Reply{}, err
	}
	return client.response()
}

// Data opens a new passive data port.
func (client *Client) Data() (io.ReadCloser, os.Error) {
	// TODO(ross): EPSV for IPv6
	reply, err := client.Do("PASV")
	if err != nil {
		return nil, err
	} else if reply.Code != 227 {
		return nil, reply
	}

	addr, err := parsePasvReply(reply.Msg)
	if err != nil {
		return nil, err
	}

	return net.DialTCP("tcp4", nil, addr)
}

var pasvRegexp = regexp.MustCompile(`([0-9]+),([0-9]+),([0-9]+),([0-9]+),([0-9]+),([0-9]+)`)

func parsePasvReply(msg string) (*net.TCPAddr, os.Error) {
	numberStrings := pasvRegexp.FindStringSubmatch(msg)
	if numberStrings == nil {
		return nil, os.NewError("PASV reply provided no port")
	}
	numbers := make([]int, len(numberStrings))
	for i, s := range numberStrings {
		numbers[i], _ = strconv.Atoi(s)
	}
	return &net.TCPAddr{
		IP:   net.IP{byte(numbers[0]), byte(numbers[1]), byte(numbers[2]), byte(numbers[3])},
		Port: numbers[4]<<8 | numbers[5],
	}, nil
}
