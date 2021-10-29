package icloud

import (
	"net/http/cookiejar"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	netscapeCookieJar "github.com/vanym/golang-netscape-cookiejar"
)

func setupCookieJar(subJar *cookiejar.Jar, path string) (*netscapeCookieJar.Jar, error) {
	opt := netscapeCookieJar.Options{
		SubJar:        subJar,
		AutoWritePath: path,
		WriteHeader:   true,
	}
	jar, err := netscapeCookieJar.New(&opt)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create cookie jar")
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
		log.Debugf("cookie file not found: %s", path)
	}
	return jar, nil
}
