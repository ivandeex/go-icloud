package icloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ivandeex/go-icloud/icloud/api"
	log "github.com/sirupsen/logrus"
)

// API endpoints
const (
	AuthEndpoint  = "https://idmsa.apple.com/appleauth/auth"
	HomeEndpoint  = "https://www.icloud.com"
	SetupEndpoint = "https://setup.icloud.com/setup/ws/1"
	DefUserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
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
	params      dict
	data        *api.StateResponse
}

// New returns API client
func NewClient(appleID, password, dataRoot, userAgent string) (*Client, error) {
	// defaults
	const (
		verify      = true
		withFamily  = true
		tempDirName = "icloud"
	)

	if userAgent == "" {
		userAgent = DefUserAgent
	}

	if dataRoot == "" {
		dataRoot = filepath.Join(os.TempDir(), tempDirName)
	}
	dataDir := filepath.Join(dataRoot, appleID)
	err := os.MkdirAll(dataRoot, 0o777)
	if err == nil {
		err = os.MkdirAll(dataDir, 0o700)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot create data directory %s: %w", dataDir, err)
	}
	cookiePath := filepath.Join(dataDir, "cookies.txt")
	sessPath := filepath.Join(dataDir, "session.txt")

	client := &http.Client{}
	jar, err := newCookieJar(cookiePath)
	if err != nil {
		return nil, fmt.Errorf("cannot load cookies from %s: %w", cookiePath, err)
	}
	client.Jar = jar

	c, err := NewClientWithOptions(client, appleID, password, userAgent, sessPath, verify, withFamily)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewClientWithOptions returns client with extra options
func NewClientWithOptions(client *http.Client, appleID, password, userAgent, sessPath string, verify, withFamily bool) (*Client, error) {
	c := &Client{
		Client:      client,
		accountName: appleID,
		password:    password,
		withFamily:  withFamily,
		sessPath:    sessPath,
		data:        &api.StateResponse{},
	}
	if err := c.session.load(sessPath); err != nil {
		return nil, fmt.Errorf("cannot load session from %s: %w", sessPath, err)
	}
	if c.session.ClientID == "" {
		c.session.ClientID = "auth-" + strings.ToLower(uuid.NewString())
	}
	return c, nil
}

// get request
func (c *Client) get(url string, res interface{}) error {
	_, err := c.request(http.MethodGet, url, nil, nil, res, false)
	return err
}

// post request
func (c *Client) post(url string, data interface{}, hdr dict, res interface{}) error {
	_, err := c.request(http.MethodPost, url, data, hdr, res, false)
	return err
}

// request will send a get/post request with retries
func (c *Client) request(method, url string, data interface{}, hdr dict, out interface{}, retried bool) ([]byte, error) {
	var (
		rd    io.Reader
		in    []byte
		inStr string
		err   error
	)
	switch d := data.(type) {
	case io.Reader:
		rd = d
		inStr = "stream"
	case []byte:
		in = d
		inStr = string(d)
	case string:
		in = []byte(d)
		inStr = d
	default:
		if d == nil {
			inStr = "null"
		} else {
			in, err = json.Marshal(d)
			inStr = string(Marshal(d))
		}
	}
	if rd == nil {
		rd = bytes.NewBuffer(in)
	}

	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	for key, val := range c.params {
		url += fmt.Sprintf("%s%s=%s", sep, key, val)
		sep = "&"
	}

	log.Tracef("%s %s %s", method, url, inStr)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		return nil, err
	}

	h := req.Header
	for k, v := range hdr {
		h.Set(k, fmt.Sprintf("%s", v))
	}
	h.Set("Origin", HomeEndpoint)
	h.Set("Referer", HomeEndpoint+"/")
	if c.userAgent != "" {
		h.Set("User-Agent", c.userAgent)
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	c.session.applyResponseHeaders(res.Header)
	if err := c.session.save(c.sessPath); err != nil {
		return nil, err
	}
	log.Tracef("Saved session in %s", c.sessPath)

	if streamPtr, wantStream := out.(*io.ReadCloser); wantStream {
		log.Tracef("streaming data from url %q", url)
		*streamPtr = res.Body
		return nil, nil
	}

	code := res.StatusCode
	strCode := strconv.Itoa(code)
	status := strings.TrimSpace(strings.TrimPrefix(res.Status, strCode))
	isAuthErr := false
	switch code {
	case 421, 450, 500:
		isAuthErr = true
	}
	body, err := io.ReadAll(res.Body)
	if err == nil {
		err = res.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	clength := res.Header.Get("Content-Length")
	if clength == "" {
		clength = strconv.Itoa(len(body)) + ".."
	}
	ctype := strings.Split(res.Header.Get("Content-Type"), ";")[0]
	isJSON := ctype == "application/json" || ctype == "text/json"
	log.Tracef("Results: code=%d noauth=%v json=%v len=%s", code, isAuthErr, isJSON, clength)

	if code >= 400 && (!isJSON || isAuthErr) {
		findmeURL, err := c.getWebserviceURL("findme")
		isFindme := err == nil && strings.Contains(url, findmeURL)
		if isFindme && !retried && code == 450 {
			log.Debug("Re-authenticating Find My iPhone service")
			if err := c.Authenticate(true, "find"); err != nil {
				log.Debug("Re-authentication failed")
			}
			return c.request(method, url, data, hdr, out, true)
		}
		if !retried && isAuthErr {
			log.Debugf("Auth error %s (%d). Retrying...", status, code)
			return c.request(method, url, data, hdr, out, true)
		}
		return nil, c.translateError(code, status, status)
	}

	if !isJSON {
		return body, nil
	}
	if err = c.decodeError(body); err != nil {
		return nil, err
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		var (
			d dict
			a []dict
			b []byte
			t string
		)
		if err = json.Unmarshal(body, &d); err == nil || d != nil {
			b = Marshal(d)
			t = "dict"
		} else if err = json.Unmarshal(body, &a); err == nil || a != nil {
			b = Marshal(a)
			t = "array"
		}
		if t != "" {
			log.Tracef("JSON %s response: %s", t, string(b))
		} else {
			log.Errorf("Invalid JSON response: %s", string(body))
		}
	}

	if out != nil {
		if err = json.Unmarshal(body, out); err != nil {
			log.Errorf("Failed to parse JSON into %T: %s", out, string(body))
			return nil, err
		}
	}
	return body, nil
}

func (c *Client) decodeError(out []byte) error {
	e := &api.ErrorResponse{}
	if err := json.Unmarshal(out, &e); err != nil || e == nil {
		return nil
	}
	var reason string
	for _, v := range []string{e.ErrorMessage, e.Reason, e.ErrorReason, e.Error} {
		if reason == "" {
			reason = v
		}
	}
	if reason == "" {
		return nil
	}
	var code int
	code = e.Code
	if code == 0 {
		code = e.ServerCode
	}
	return c.translateError(code, "", reason)
}

func (c *Client) translateError(code int, status string, reason string) error {
	if c.Requires2SA() && reason == "Missing X-APPLE-WEBAUTH-TOKEN cookie" {
		return Err2SARequired
	}
	switch status {
	case "ZONE_NOT_FOUND", "AUTHENTICATION_FAILED":
		reason = "Please log into https://icloud.com/ to manually finish setting up your iCloud service"
		return NewErrAPI(code, status, reason, false)
	case "ACCESS_DENIED":
		reason += ".  Please wait a few minutes then try again"
		reason += ". The remote servers might be trying to throttle requests."
	}
	switch code {
	case 421, 450, 500:
		reason = "Authentication required for Account."
	}
	return NewErrAPI(code, status, reason, false)
}
