package icloud

import (
	"errors"
	"fmt"
	"strings"
)

type ErrApple error

// NewErr returns generic Apple iCloud error
func NewErr(msg string) ErrApple {
	return ErrApple(errors.New(msg))
}

// ErrAPI is subclass of API related iCloud errors
type ErrAPI struct {
	ErrApple
	Code  int
	Retry bool
}

// NewErrAPIResponse returns new API related iCloud error
func NewErrAPI(code int, status string, reason string, retry bool) ErrAPI {
	msg := reason
	if status != "" {
		if msg == "" {
			msg = status
		} else {
			if !strings.HasSuffix(msg, ".") {
				msg += "."
			}
			msg += " " + status
		}
	}
	if code != 0 {
		msg = fmt.Sprintf("%s (%d)", msg, code)
	}
	if retry {
		msg += ". Retrying ..."
	}
	return ErrAPI{ErrApple: NewErr(msg), Code: code, Retry: retry}
}

var (
	ErrServiceNotActive  = NewErr("icloud service not activated")
	ErrLoginFailed       = NewErr("icloud login failed")
	Err2SARequired       = NewErr("2-step authentication required for account")
	ErrNoStoredPassword  = NewErr("no stored icloud password available")
	ErrNoDevices         = NewErr("no icloud device")
	ErrWrongVerification = NewErr("wrong verification code")
	ErrNotFound          = NewErr("path not found")
	ErrNotDir            = NewErr("path is not a directory")
	ErrNotFile           = NewErr("path is not a file")
)
