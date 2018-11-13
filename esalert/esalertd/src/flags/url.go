package flags

import (
	"net/url"
	"github.com/tehmoon/errors"
	"fmt"
)

func derivePublicURL(public, listen string) (u string, err error) {
	publicURL, err := url.Parse(public)
	if err != nil {
		return "", errors.Wrapf(err, "Fail to assert flag %q", "public-url")
	}

	listenURL := url.URL{
		Host: listen,
	}

	if listenURL.Port() == "" {
		return "", errors.Errorf("Missing port number in %q flag", "listen")
	}

	if listenURL.Hostname() != "" {
		publicURL.Host = listenURL.Hostname()
	}

	publicURL.Host = fmt.Sprintf("%s:%s", publicURL.Hostname(), listenURL.Port())

	return publicURL.String(), nil
}
