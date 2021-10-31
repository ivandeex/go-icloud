package icloud

import (
	log "github.com/sirupsen/logrus"
)

// Authenticate handles authentication, and persists cookies so that
// subsequent logins will not cause additional e-mails from Apple.
func (c *Client) Authenticate(force_refresh bool, service string) (err error) {
	success := false

	if c.session.SessionToken != "" && !force_refresh {
		c.data, err = c.validateToken()
		if err != nil {
			log.Debugf("Invalid authentication token, will log in from scratch: %v", err)
		} else {
			success = true
		}
	}

	if !success && service != "" {
		// TODO
		appsRaw := c.data["apps"]
		apps := appsRaw.(dict)
		appRaw := apps[service]
		app := appRaw.(dict)
		canLaunchWithOneFactorRaw := app["canLaunchWithOneFactor"]
		canLaunchWithOneFactor := canLaunchWithOneFactorRaw.(bool)
		if canLaunchWithOneFactor {
			log.Debugf("Authenticating as %s for %s", c.accountName, service)
			err = c.authenticateWithCredentialsService(service)
			if err != nil {
				log.Debugf("Could not log into service. Attempting brand new login.")
			} else {
				success = true
			}
		}
	}

	if !success {
		log.Debugf("Authenticating as %s", c.accountName)

		trustTokens := []string{}
		if c.session.TrustToken != "" {
			trustTokens = append(trustTokens, c.session.TrustToken)
		}

		data := dict{
			"accountName": c.accountName,
			"password":    c.password,
			"rememberMe":  true,
			"trustTokens": trustTokens,
		}

		hdr := c.getAuthHeaders(nil)
		if c.session.SCnt != "" {
			hdr["scnt"] = c.session.SCnt
		}
		if c.session.SessionID != "" {
			hdr["X-Apple-ID-Session-Id"] = c.session.SessionID
		}

		if err = c.post(AuthEndpoint+"/signin?isRememberMeEnabled=true", data, hdr, nil); err != nil {
			return ErrLoginFailed // "Invalid email/password combination."
		}

		if err = c.authenticateWithToken(); err != nil {
			return err
		}
	}

	log.Debugf("Authentication completed successfully")
	return nil
}

func (c *Client) authenticateWithToken() error {
	data := dict{
		"accountCountryCode": c.session.AccountCountry,
		"dsWebAuthToken":     c.session.SessionToken,
		"extended_login":     true,
		"trustToken":         c.session.TrustToken,
	}
	var res dict
	if err := c.post(SetupEndpoint+"/accountLogin", data, nil, &res); err != nil {
		return ErrLoginFailed
	}
	c.data = res
	return nil
}

// Authenticate to a specific service using credentials.
func (c *Client) authenticateWithCredentialsService(service string) (err error) {
	data := dict{
		"appName":  service,
		"apple_id": c.accountName,
		"password": c.password,
	}
	if err = c.post(SetupEndpoint+"/accountLogin", data, nil, nil); err != nil {
		return ErrLoginFailed
	}
	c.data, err = c.validateToken()
	return err
}

// validateToken checks if the current access token is still valid.
func (c *Client) validateToken() (dict, error) {
	log.Debugf("Checking session token validity")
	var res dict
	err := c.post(SetupEndpoint+"/validate", nil, nil, &res)
	log.Debugf("validateToken got err:%v res:%s", err, Marshal(res))
	if err != nil {
		log.Debugf("Invalid authentication token")
		return nil, err
	}
	log.Debugf("Session token is still valid")
	return res, nil
}

func (c *Client) getAuthHeaders(overrides dict) dict {
	h := dict{
		"Accept":                           "*/*",
		"Content-Type":                     "application/json",
		"X-Apple-OAuth-Client-Id":          "d39ba9916b7251055b22c7f910e2ea796ee65e98b2ddecea8f5dde8d9d1a815d",
		"X-Apple-OAuth-Client-Type":        "firstPartyAuth",
		"X-Apple-OAuth-Redirect-URI":       "https://www.icloud.com",
		"X-Apple-OAuth-Require-Grant-Code": "true",
		"X-Apple-OAuth-Response-Mode":      "web_message",
		"X-Apple-OAuth-Response-Type":      "code",
		"X-Apple-OAuth-State":              c.session.ClientID,
		"X-Apple-Widget-Key":               "d39ba9916b7251055b22c7f910e2ea796ee65e98b2ddecea8f5dde8d9d1a815d",
	}
	for k, v := range overrides {
		h[k] = v
	}
	return h
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

func (c *Client) getWebserviceURL(service string) string {
	return "TODO!"
}
