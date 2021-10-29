package icloud

import (
	"github.com/pkg/errors"
)

func (c *Client) getWebserviceURL(service string) string {
	return "TODO!"
}

func (c *Client) Authenticate(flag bool, service string) error {
	return errors.New("not implemented")
}

// Requires2FA returns true if 2-factor auth is required
func (c *Client) Requires2FA() bool {
	return false
}

// Requires2SA returns true if 2-step auth is required
func (c *Client) Requires2SA() bool {
	return false
}

func (c *Client) Verify2FACode(code string) (bool, error) {
	return false, nil
}

// Device describes a user device like iPhone, iPad and so on
type Device struct {
	DeviceType  string
	PhoneNumber string
}

// TrustedDevices returns slice of trusted devices
func (c *Client) TrustedDevices() ([]*Device, error) {
	return nil, nil
}

// SendVerificationCode makes iCloud send verification code to a device
func (c *Client) SendVerificationCode(d *Device) error {
	return nil
}

// ValidateVerificationCode received on a device
func (c *Client) ValidateVerificationCode(d *Device, code string) (bool, error) {
	return false, nil
}
