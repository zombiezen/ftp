// Copyright (c) 2011 Ross Light.

// Package ftp provides a minimal FTP client.
package ftp

import (
	"strconv"
	"strings"
)

// Reply is a response from a server.  This may also be used as an error.
type Reply struct {
	Code int
	Msg  string
}

// Positive returns whether the reply is positive.
func (r Reply) Positive() bool {
	return r.Code < 400
}

func (r Reply) String() string {
	lines := strings.Split(r.Msg, "\n")
	if len(lines) > 1 {
		lines[0] = strconv.Itoa(r.Code) + "-" + lines[0]
		lines[len(lines)-1] = strconv.Itoa(r.Code) + " " + lines[len(lines)-1]
		return strings.Join(lines, "\r\n")
	}
	return strconv.Itoa(r.Code) + " " + r.Msg
}
