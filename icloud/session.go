package icloud

import (
	"encoding/json"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// sessionData keeps session data
type sessionData struct {
	ClientID       string
	AccountCountry string
	SessionID      string
	SessionToken   string
	TrustToken     string
	SCnt           string
}

func (s sessionData) save(path string) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (s *sessionData) load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil && os.IsNotExist(err) {
		log.Debugf("session file not found: %s", path)
		return nil
	}
	if err == nil {
		err = json.Unmarshal(data, s)
	}
	return err
}

func (s *sessionData) applyResponseHeader(h http.Header) {
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
