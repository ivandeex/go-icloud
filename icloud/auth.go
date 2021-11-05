package icloud

import (
	"errors"

	"github.com/ivandeex/go-icloud/icloud/api"
	log "github.com/sirupsen/logrus"
)

// Authenticate handles authentication, and persists cookies so that
// subsequent logins will not cause additional e-mails from Apple.
func (c *Client) Authenticate(force_refresh bool, service string) (err error) {
	success := false

	if c.session.SessionToken != "" && !force_refresh {
		c.data, err = c.validateToken()
		if err != nil {
			log.Debugf("Will log in from scratch: %v", err)
		} else {
			success = true
		}
	}

	if !success && service != "" {
		if allows1F, _ := c.data.Apps.AllowsOneFactor(service); allows1F {
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

		hdr := c.getAuthHeaders(true)

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
	var res *api.StateResponse
	if err := c.post(SetupEndpoint+"/accountLogin", data, nil, &res); err != nil {
		return ErrLoginFailed
	}
	c.data = res
	return nil
}

// Authenticate to a specific service using credentials.
func (c *Client) authenticateWithCredentialsService(service string) error {
	data := dict{
		"appName":  service,
		"apple_id": c.accountName,
		"password": c.password,
	}
	err := c.post(SetupEndpoint+"/accountLogin", data, nil, nil)
	if err != nil {
		return ErrLoginFailed
	}
	c.data, err = c.validateToken()
	if err != nil {
		return err
	}
	return nil
}

// validateToken checks if the current access token is still valid.
func (c *Client) validateToken() (*api.StateResponse, error) {
	log.Debugf("Checking session token validity")
	var res *api.StateResponse
	if err := c.post(SetupEndpoint+"/validate", nil, nil, &res); err != nil {
		log.Debugf("Invalid authentication token: %v", err)
		return nil, err
	}
	log.Debugf("Session token is still valid")
	return res, nil
}

// getAuthHeaders returns actual authentication headers
func (c *Client) getAuthHeaders(useSession bool) dict {
	const OauthKey = "d39ba9916b7251055b22c7f910e2ea796ee65e98b2ddecea8f5dde8d9d1a815d"
	h := dict{}
	h["Accept"] = "*/*"
	h["Content-Type"] = "application/json"
	h["X-Apple-Widget-Key"] = OauthKey
	h["X-Apple-OAuth-Client-Id"] = OauthKey
	h["X-Apple-OAuth-Client-Type"] = "firstPartyAuth"
	h["X-Apple-OAuth-Redirect-URI"] = "https://www.icloud.com"
	h["X-Apple-OAuth-Require-Grant-Code"] = "true"
	h["X-Apple-OAuth-Response-Mode"] = "web_message"
	h["X-Apple-OAuth-Response-Type"] = "code"
	h["X-Apple-OAuth-State"] = c.session.ClientID
	if useSession && c.session.SCnt != "" {
		h["scnt"] = c.session.SCnt
	}
	if useSession && c.session.SessionID != "" {
		h["X-Apple-ID-Session-Id"] = c.session.SessionID
	}
	return h
}

// Requires2FA returns true if 2-factor authentication is required.
func (c *Client) Requires2FA() bool {
	return c.data.DsInfo.HsaVersion == 2 && (c.data.HsaChallengeRequired || !c.IsTrustedSession())
}

// Requires2SA returns true if 2-step authentication is required.
func (c *Client) Requires2SA() bool {
	return c.data.DsInfo.HsaVersion >= 1 && (c.data.HsaChallengeRequired || !c.IsTrustedSession())
}

// Validate2FACode verifies a code received via Apple's 2FA system (HSA2).
func (c *Client) Validate2FACode(code string) error {
	data := dict{
		"securityCode": dict{
			"code": code,
		},
	}
	hdr := c.getAuthHeaders(true)
	hdr["Accept"] = "application/json"
	err := c.post(AuthEndpoint+"/verify/trusteddevice/securitycode", data, hdr, nil)
	if err != nil {
		if apiErr, ok := err.(ErrAPI); ok {
			if apiErr.Code == api.CodeWrongVerification2 {
				return ErrWrongVerification
			}
		}
	}
	return err
}

// IsTrustedSession returns true if current session is trusted.
func (c *Client) IsTrustedSession() bool {
	return c.data.HsaTrustedBrowser
}

// TrustSession requests session trust to avoid user log in going forward.
func (c *Client) TrustSession() error {
	hdr := c.getAuthHeaders(true)
	err := c.post(AuthEndpoint+"/2sv/trust", nil, hdr, nil)
	if err == nil {
		err = c.authenticateWithToken()
	}
	return err
}

// TrustedDevices returns slice of trusted devices
func (c *Client) TrustedDevices() ([]api.Device, error) {
	var res *api.DeviceResponse
	if err := c.get(SetupEndpoint+"/listDevices", &res); err != nil || res == nil {
		return nil, errors.New("invalid response from listDevices")
	}
	if len(res.Devices) == 0 {
		return nil, ErrNoDevices
	}
	return res.Devices, nil
}

// SendVerificationCode makes iCloud send verification code to a device
func (c *Client) SendVerificationCode(dev *api.Device) error {
	var res *api.SuccessResponse
	if err := c.post(SetupEndpoint+"/sendVerificationCode", dev, nil, &res); err != nil {
		return err
	}
	if res == nil || !res.Success {
		return errors.New("failed to send verification code")
	}
	return nil
}

// ValidateVerificationCode received on a device
func (c *Client) ValidateVerificationCode(dev *api.Device, code string) error {
	d := dev.Dict()
	d["verificationCode"] = code
	d["trustBrowser"] = true
	if err := c.post(SetupEndpoint+"/validateVerificationCode", d, nil, nil); err != nil {
		if apiErr, ok := err.(ErrAPI); ok {
			if apiErr.Code == api.CodeWrongVerification {
				return ErrWrongVerification
			}
		}
		return err
	}
	if err := c.TrustSession(); err != nil {
		if apiErr, ok := err.(ErrAPI); ok {
			if apiErr.Code == api.CodeNotFound {
				log.Infof("You seem to lack trusted Apple devices. Authenticating again...")
				err = c.Authenticate(false, "")
			}
		}
		return err
	}
	if c.Requires2SA() {
		return ErrLoginFailed
	}
	return nil
}
