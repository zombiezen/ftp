package ftp

import (
	"strconv"
	"strings"
)

type Reply struct {
	Code int
	Msg  string
}

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
