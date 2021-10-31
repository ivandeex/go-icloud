package icloud

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrICloud error

// NewErr returns generic iCloud error
func NewErr(msg string) ErrICloud {
	return ErrICloud(errors.New(msg))
}

// ErrCloudAPIResponse is subclass of API related iCloud errors
type ErrAPIResponse struct {
	ErrICloud
	Retry bool
}

// NewErrAPIResponse returns new API related iCloud error
func NewErrAPIResponse(code int, status string, reason string, retry bool) ErrAPIResponse {
	msg := reason
	if status != "" {
		msg = fmt.Sprintf("%s (%s)", msg, status)
	} else if code != 0 {
		msg = fmt.Sprintf("%s (%d)", msg, code)
	}
	if retry {
		msg += ". Retrying ..."
	}
	return ErrAPIResponse{ErrICloud: NewErr(msg), Retry: retry}
}

var (
	ErrServiceNotActive = NewErr("iCloud service not activated")
	ErrLoginFailed      = NewErr("iCloud login failed")
	Err2SARequired      = NewErr("2-step authentication required for account")
	ErrNoStoredPassword = NewErr("no stored iCloud password available")
	ErrNoDevices        = NewErr("no iCloud device")
)
