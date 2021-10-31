package icloud

import (
	"encoding/json"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// sessionData keeps session data
type sessionData struct {
	ClientID       string `json:"client_id"`
	AccountCountry string `json:"account_country"`
	SessionID      string `json:"session_id"`
	SessionToken   string `json:"session_token"`
	TrustToken     string `json:"trust_token"`
	SCnt           string `json:"scnt"`
}

func (s sessionData) save(path string) error {
	return os.WriteFile(path, Marshal(s), 0600)
}

func (s *sessionData) load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil && os.IsNotExist(err) {
		log.Debugf("Session file not found: %s", path)
		return nil
	}
	if err == nil {
		err = json.Unmarshal(data, s)
	}
	return err
}

func (s *sessionData) applyResponseHeaders(h http.Header) {
	var v string
	if v = h.Get("X-Apple-ID-Account-Country"); v != "" {
		s.AccountCountry = v
	}
	if v = h.Get("X-Apple-ID-Session-Id"); v != "" {
		s.SessionID = v
	}
	if v = h.Get("X-Apple-Session-Token"); v != "" {
		s.SessionToken = v
	}
	if v = h.Get("X-Apple-TwoSV-Trust-Token"); v != "" {
		s.TrustToken = v
	}
	if v = h.Get("scnt"); v != "" {
		s.SCnt = v
	}
}
