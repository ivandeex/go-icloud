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
	data        dict
}

type dict map[string]interface{}

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
	jar, err := newCookieJar(cookiePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load cookies from %s", cookiePath)
	}
	client.Jar = jar

	c, err := NewWithOptions(client, appleID, password, userAgent, sessPath, verify, withFamily)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewWithOptions returns client with extra options
func NewWithOptions(client *http.Client, appleID, password, userAgent, sessPath string, verify, withFamily bool) (*Client, error) {
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

// post request
func (c *Client) post(url string, data dict, hdr dict, out interface{}) error {
	_, err := c.request(http.MethodPost, url, data, hdr, out, false)
	return err
}

// request will send a get/post request with retries
func (c *Client) request(method, url string, data dict, hdr dict, out interface{}, retried bool) ([]byte, error) {
	log.Debugf("%s %s %s", method, url, Marshal(data))
	var (
		dataIn []byte
		err    error
	)
	if data != nil {
		if dataIn, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(dataIn))
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
	log.Tracef("---\n dataIn: %s\n request: %v\n err: %v\n response: %v\n---", dataIn, req, err, res)
	if err != nil {
		return nil, err
	}

	c.session.applyResponseHeaders(res.Header)
	if err := c.session.save(c.sessPath); err != nil {
		return nil, err
	}
	log.Debugf("Saved session data to file %s", c.sessPath)
	log.Debugf("Cookies saved aitomatically")

	code := res.StatusCode
	strCode := strconv.Itoa(code)
	status := strings.TrimSpace(strings.TrimPrefix(res.Status, strCode))
	isAuthErr := false
	switch code {
	case 421, 450, 500:
		isAuthErr = true
	}
	OK := code < 400
	clength := res.Header.Get("Content-Length")
	dataOut, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	_ = res.Body.Close()
	if clength == "" {
		clength = strconv.Itoa(len(dataOut)) + "."
	}
	ctype := strings.Split(res.Header.Get("Content-Type"), ";")[0]
	isJSON := ctype == "application/json" || ctype == "text/json"

	log.Debugf("Results: OK=%v AuthErr=%v IsJSON=%v Code=%d Length=%s",
		OK, isAuthErr, isJSON, code, clength)

	if !OK && (!isJSON || isAuthErr) {
		isFindme := strings.Contains(url, c.getWebserviceURL("findme"))
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
		return nil, c.translateErrors(code, status, status)
	}

	if !isJSON {
		return dataOut, nil
	}

	if err = c.handleReason(dataOut); err != nil {
		return nil, err
	}
	if out != nil {
		err = json.Unmarshal(dataOut, &out)
		if err != nil || out == nil {
			log.Warn("Failed to parse JSON response")
			return nil, err
		}
		log.Debugf("json response: %s", string(Marshal(out)))
	}
	return dataOut, nil
}

func (c *Client) handleReason(dataOut []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(dataOut, &data); err != nil || data == nil {
		return nil
	}
	out := map[string]string{}
	for k, v := range data {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}

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
