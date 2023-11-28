package client

import (
	"net/http"
	"time"
)

var defaultTimeout = time.Duration(30 * time.Second)
var defaultCookieJar = http.DefaultClient.Jar

type httpClientOptions struct {
	timeout *time.Duration
	jar     http.CookieJar
}

func NewHTTPClient(options ...HTTPClientOption) (*http.Client, error) {
	var err error

	opts := httpClientOptions{
		timeout: &defaultTimeout,
		jar:     defaultCookieJar,
	}

	for _, option := range options {
		if err = option(&opts); err != nil {
			return nil, err
		}
	}

	c := &http.Client{
		Jar:     opts.jar,
		Timeout: *opts.timeout,
	}

	return c, nil
}

type HTTPClientOption func(options *httpClientOptions) error

func WithTimeout(t time.Duration) HTTPClientOption {
	return func(options *httpClientOptions) error {
		*options.timeout = t

		return nil
	}
}

func WithCookieJar(j http.CookieJar) HTTPClientOption {
	return func(options *httpClientOptions) error {
		if j != nil {
			options.jar = j
		}

		return nil
	}
}
