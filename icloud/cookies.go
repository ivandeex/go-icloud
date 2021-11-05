package icloud

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"

	log "github.com/sirupsen/logrus"
	netscapeCookieJar "github.com/vanym/golang-netscape-cookiejar"
)

func newCookieJar(path string) (http.CookieJar, error) {
	baseJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create basic cookie jar: %w", err)
	}
	opt := netscapeCookieJar.Options{
		SubJar:        baseJar,
		AutoWritePath: path,
		WriteHeader:   true,
	}
	jar, err := netscapeCookieJar.New(&opt)
	if err != nil {
		return nil, fmt.Errorf("cannot create on-disk cookie jar: %w", err)
	}
	file, err := os.Open(path)
	if err == nil {
		_, err = jar.ReadFrom(file)
		_ = file.Close()
	}
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		log.Debugf("Cookie file not found: %s", path)
	}
	return jar, nil
}
