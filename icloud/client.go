package icloud

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// API endpoints
const (
	AuthEndpoint  = "https://idmsa.apple.com/appleauth/auth"
	HomeEndpoint  = "https://www.icloud.com"
	SetupEndpoint = "https://setup.icloud.com/setup/ws/1"
)

// Client is iCloud API client
type Client struct {
	Client      *http.Client
	userAgent   string
	accountName string
	password    string
	withFamily  bool
	session     sessionData
	sessPath    string
}

// New returns API client
func New(appleID, password string) (*Client, error) {
	// defaults
	const (
		verify      = true
		withFamily  = true
		tempDirName = "icloud"
		userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
	)

	dataRoot := filepath.Join(os.TempDir(), tempDirName)
	dataDir := filepath.Join(dataRoot, appleID)
	err := os.MkdirAll(dataRoot, 0o777)
	if err == nil {
		err = os.MkdirAll(dataDir, 0o700)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create data directory %s", dataDir)
	}
	cookiePath := filepath.Join(dataDir, "cookies.txt")
	sessPath := filepath.Join(dataDir, "session.txt")

	client := &http.Client{}
	jar, err := setupCookieJar(nil, cookiePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load cookies from %s", cookiePath)
	}
	client.Jar = jar

	c, err := NewClient(client, appleID, password, userAgent, sessPath, verify, withFamily)
	if err == nil {
		err = c.Authenticate(false, "")
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewClient returns client with extra options
func NewClient(client *http.Client, appleID, password, userAgent, sessPath string, verify, withFamily bool) (*Client, error) {
	c := &Client{
		Client:      client,
		accountName: appleID,
		password:    password,
		withFamily:  withFamily,
		sessPath:    sessPath,
	}
	if err := c.session.load(sessPath); err != nil {
		return nil, errors.Wrapf(err, "cannot load session from %s", sessPath)
	}
	if c.session.ClientID == "" {
		c.session.ClientID = "auth-" + strings.ToLower(uuid.NewString())
	}
	return c, nil
}

func (c *Client) Request(method, url string, data interface{}, retried bool) (*http.Response, error) {
	log.Debugf("%s %s %q", method, url, data)
	dataIn, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(dataIn))
	if err != nil {
		return nil, err
	}

	h := req.Header
	h.Set("Origin", HomeEndpoint)
	h.Set("Referer", HomeEndpoint+"/")
	if c.userAgent != "" {
		h.Set("User-Agent", c.userAgent)
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	c.session.applyResponseHeader(res.Header)
	if err := c.session.save(c.sessPath); err != nil {
		return nil, err
	}
	log.Debugf("Saved session data to file %s", c.sessPath)
	// Cookies has been saved to file aitomatically

	contentType := strings.Split(res.Header.Get("Content-Type"), ";")[0]
	isJSON := contentType == "application/json" || contentType == "text/json"

	ok := res.StatusCode < 400
	isAuthErr := false
	switch res.StatusCode {
	case 421, 450, 500:
		isAuthErr = true
	}

	if !ok && (!isJSON || isAuthErr) {
		isFindme := strings.Contains(url, c.getWebserviceURL("findme"))
		if isFindme && !retried && res.StatusCode == 450 {
			log.Debug("Re-authenticating Find My iPhone service")
			if err := c.Authenticate(true, "find"); err != nil {
				log.Debug("Re-authentication failed")
			}
			return c.Request(method, url, data, true)
		}
		if !retried && isAuthErr {
			log.Debugf("Auth error %s (%d)", res.Status, res.StatusCode)
			return c.Request(method, url, data, true)
		}

		return nil, c.translateErrors(res.StatusCode, res.Status, res.Status)
	}

	if !isJSON {
		return res, nil
	}

	dataOut, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	_ = res.Body.Close()
	var out map[string]string
	err = json.Unmarshal(dataOut, &out)
	if err != nil || out == nil {
		log.Warn("Failed to parse JSON response")
		return nil, err
	}

	log.Debugf("json response: %#v", out)
	err = c.handleReason(out)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) handleReason(out map[string]string) error {
	var reason string
	if reason == "" {
		reason = out["errorMessage"]
	}
	if reason == "" {
		reason = out["reason"]
	}
	if reason == "" {
		reason = out["errorReason"]
	}
	if reason == "" {
		reason = out["error"]
	}
	//if (reason is not string) { reason = "Unknown reason" }
	if reason == "" {
		return nil
	}

	var status string
	status = out["errorCode"]
	if status == "" {
		status = out["serverErrorCode"]
	}
	var code int
	code, err := strconv.Atoi(status)
	if err == nil {
		status = ""
	}
	return c.translateErrors(code, status, reason)
}

func (c *Client) translateErrors(code int, status string, reason string) error {
	if c.Requires2SA() && reason == "Missing X-APPLE-WEBAUTH-TOKEN cookie" {
		return Err2SARequired
	}
	switch status {
	case "ZONE_NOT_FOUND", "AUTHENTICATION_FAILED":
		reason = "Please log into https://icloud.com/ to manually finish setting up your iCloud service"
		return NewErrAPIResponse(code, status, reason, false)
	case "ACCESS_DENIED":
		reason += ".  Please wait a few minutes then try again"
		reason += ". The remote servers might be trying to throttle requests."
	}
	switch code {
	case 421, 450, 500:
		reason = "Authentication required for Account."
	}
	return NewErrAPIResponse(code, status, reason, false)
}
