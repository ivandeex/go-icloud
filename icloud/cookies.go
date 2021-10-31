package icloud

import (
	"net/http/cookiejar"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	netscapeCookieJar "github.com/vanym/golang-netscape-cookiejar"
)

func newCookieJar(path string) (*netscapeCookieJar.Jar, error) {
	baseJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create basic cookie jar")
	}
	opt := netscapeCookieJar.Options{
		SubJar:        baseJar,
		AutoWritePath: path,
		WriteHeader:   true,
	}
	jar, err := netscapeCookieJar.New(&opt)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create netscape cookie jar")
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
