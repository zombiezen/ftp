// Copyright (c) 2011 Ross Light.

package ftp

import (
	"strconv"
	"strings"
)

// Code is an FTP reply code.
type Code int

// FTP reply codes defined in RFC 959.
const (
	CodeRestartMarker    Code = 110
	CodeServiceReadySoon      = 120
	CodeStartingTransfer      = 125
	CodeFileStatusOkay        = 150

	CodeOkay            = 200
	CodeSuperfluous     = 202
	CodeSystemStatus    = 211
	CodeDirectoryStatus = 212
	CodeFileStatus      = 213
	CodeHelpMessage     = 214
	CodeSystemType      = 215
	CodeServiceReady    = 220
	CodeServiceClosing  = 221
	CodeNoTransfer      = 225
	CodeClosingData     = 226
	CodePassive         = 227
	CodeLoggedIn        = 230
	CodeActionOkay      = 250
	CodeCreated         = 257

	CodeNeedPassword       = 331
	CodeNeedAccount        = 332
	CodePendingInformation = 350

	CodeServiceNotAvailable = 421
	CodeCantOpenData        = 425
	CodeTransferAborted     = 426
	CodeActionNotTaken      = 450
	CodeLocalError          = 451
	CodeInsufficientStorage = 452

	CodeUnrecognizedCommand     = 500
	CodeParameterSyntaxError    = 501
	CodeNotImplemented          = 502
	CodeBadSequence             = 503
	CodeParameterNotImplemented = 504
	CodeNotLoggedIn             = 530
	CodeNoAccount               = 532
	CodeFileUnavailable         = 550
	CodePageTypeUnknown         = 551
	CodeExceededQuota           = 552
	CodeFileNameNotAllowed      = 553
)

// Preliminary returns whether the code indicates a preliminary positive reply.
func (code Code) Preliminary() bool {
	return code/100 == 1
}

// Positive returns whether the code is positive.
func (code Code) Positive() bool {
	firstDigit := code / 100
	return firstDigit == 1 || firstDigit == 2 || firstDigit == 3
}

// Complete returns whether this code indicates a complete reply.  A complete reply code is not
// necessarily positive.
func (code Code) Complete() bool {
	firstDigit := code / 100
	return firstDigit == 2 || firstDigit == 4 || firstDigit == 5
}

// PositiveComplete returns whether this code is a positive completion.
func (code Code) PositiveComplete() bool {
	return code/100 == 2
}

// Temporary returns whether this code indicates a temporary error.
func (code Code) Temporary() bool {
	return code/100 == 4
}

func (code Code) String() string {
	return strconv.Itoa(int(code))
}

// Reply is a response from a server.  This may also be used as an error.
type Reply struct {
	Code
	Msg string
}

func (r Reply) String() string {
	lines := strings.Split(r.Msg, "\n")
	if len(lines) > 1 {
		lines[0] = r.Code.String() + "-" + lines[0]
		lines[len(lines)-1] = r.Code.String() + " " + lines[len(lines)-1]
		return strings.Join(lines, "\r\n")
	}
	return r.Code.String() + " " + r.Msg
}
